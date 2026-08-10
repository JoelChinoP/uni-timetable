<script lang="ts">
	import { onMount } from 'svelte';

	import TopBar from '../components/TopBar.svelte';
	import CourseExplorer from '../components/CourseExplorer.svelte';
	import WeeklyPlanner from '../components/WeeklyPlanner.svelte';
	import CourseDetailModal from '../components/CourseDetailModal.svelte';
	import Modal from '../components/panel/Modal.svelte';
	import { getDashboard, createSharedTimetable } from '../api/planner';
	import { selection, toggleGroup, clearSelection, pruneSelection } from '../stores/selection';
	import {
		buildPlannerEvents,
		buildPlannerSummary,
		getSelectedCourseGroups,
		groupRelatedCourses,
		matchesCourseSearch,
	} from '../utils/planner';
	import { downloadICalendar, renderSchedulePng } from '../utils/scheduleExport';
	import { buildICalendar } from '../utils/calendarResources';
	import type { Course, PlannerConflict, PlannerData, PlannerEvent } from '../types/planner';
	import type { AuthUser } from '../types/auth';

	export let user: AuthUser | null = null;
	export let busy = false;
	export let onNavigate: (path: string) => void;
	export let onLogout: () => void;

	let data: PlannerData | null = null;
	let loadError = '';
	let searchQuery = '';
	let selectedYears: number[] = [];
	let detailCourse: Course | null = null;
	let focusedSessionId: number | null = null;
	let sharing = false;
	let shareFeedback = '';
	let shareUrl = '';
	let mobileView: 'courses' | 'planner' = 'courses';
	let events: PlannerEvent[] = [];
	let conflicts: PlannerConflict[] = [];
	let imagePreviewUrl = '';
	let calendarModalOpen = false;
	let exportError = '';
	let shareTimer: ReturnType<typeof setTimeout> | undefined;
	const YEAR_FILTER_STORAGE_KEY = 'uni-timetable:year-filter:v1';
	const localDate = (date: Date) =>
		new Date(date.getTime() - date.getTimezoneOffset() * 60_000).toISOString().slice(0, 10);
	const today = new Date();
	let calendarStart = localDate(today);
	let calendarEnd = localDate(new Date(today.getTime() + 120 * 24 * 60 * 60 * 1000));

	async function load() {
		loadError = '';
		try {
			data = await getDashboard();
			pruneSelection(data.courses);
		} catch {
			loadError = 'No se pudo cargar el horario de cursos.';
		}
	}

	function loadSelectedYears() {
		try {
			const value = JSON.parse(window.localStorage.getItem(YEAR_FILTER_STORAGE_KEY) ?? '[]');
			return Array.isArray(value)
				? [
						...new Set(value.filter((year): year is number => Number.isInteger(year) && year > 0)),
					].sort()
				: [];
		} catch {
			return [];
		}
	}

	function setSelectedYears(years: number[]) {
		selectedYears = years;
		try {
			window.localStorage.setItem(YEAR_FILTER_STORAGE_KEY, JSON.stringify(years));
		} catch {
			// ponytail: el filtro sigue funcionando aunque el navegador bloquee el almacenamiento.
		}
	}

	function clearShareTimer() {
		if (shareTimer) clearTimeout(shareTimer);
		shareTimer = undefined;
	}

	function dismissShare() {
		clearShareTimer();
		shareFeedback = '';
		shareUrl = '';
	}

	function scheduleShareDismissal() {
		clearShareTimer();
		shareTimer = setTimeout(dismissShare, 10_000);
	}

	onMount(() => {
		selectedYears = loadSelectedYears();
		void load();
		return clearShareTimer;
	});

	$: allCourses = data?.courses ?? [];
	$: availableYears = [...new Set(allCourses.map((course) => course.academicYear))].sort();
	$: visibleBundles = groupRelatedCourses(allCourses).filter((bundle) => {
		const related = [...(bundle.theory ? [bundle.theory] : []), ...bundle.laboratories];
		return (
			(selectedYears.length === 0 ||
				related.some((course) => selectedYears.includes(course.academicYear))) &&
			related.some((course) => matchesCourseSearch(course, searchQuery))
		);
	});
	$: filteredCourses = visibleBundles.flatMap((bundle) => [
		...(bundle.theory ? [bundle.theory] : []),
		...bundle.laboratories,
	]);
	$: selectedCourseGroups = getSelectedCourseGroups(allCourses, $selection);
	$: ({ events, conflicts } = buildPlannerEvents(selectedCourseGroups));
	$: summary = buildPlannerSummary(selectedCourseGroups, conflicts);
	$: conflictingCourseIds = new Set(
		events.filter((event) => event.conflictIds.length > 0).map((event) => event.courseId),
	);

	function openEvent(event: PlannerEvent) {
		const course = allCourses.find(({ id }) => id === event.courseId) ?? null;
		if (course) {
			detailCourse = course;
			focusedSessionId = event.sessionId;
		}
	}

	async function shareBoard() {
		if (sharing || summary.selectedCourses === 0) return;
		sharing = true;
		dismissShare();
		try {
			const { id } = await createSharedTimetable($selection);
			shareUrl = `${window.location.origin}/s/${id}`;
			try {
				await navigator.clipboard.writeText(shareUrl);
				shareFeedback = 'Enlace copiado';
			} catch {
				shareFeedback = 'Enlace listo para copiar';
			}
			scheduleShareDismissal();
		} catch (error) {
			shareFeedback = error instanceof Error ? error.message : 'No se pudo crear el enlace';
			shareUrl = '';
		} finally {
			sharing = false;
		}
	}

	async function copyShareUrl() {
		try {
			await navigator.clipboard.writeText(shareUrl);
			shareFeedback = 'Enlace copiado';
		} catch {
			shareFeedback = 'Selecciona el enlace y cópialo manualmente';
		}
	}

	function previewImage() {
		if (!user || !data || events.length === 0) return;
		try {
			imagePreviewUrl = renderSchedulePng(
				events,
				data.academicHours,
				data.termLabel,
				document.documentElement.dataset.theme === 'dark',
			);
		} catch (error) {
			shareFeedback = error instanceof Error ? error.message : 'No se pudo generar la imagen';
		}
	}

	function exportCalendar() {
		if (!user || !data) return;
		exportError = '';
		if (!calendarStart || !calendarEnd || calendarStart > calendarEnd) {
			exportError = 'Selecciona un rango de fechas válido.';
			return;
		}
		const calendar = buildICalendar(
			events,
			data.academicHours,
			calendarStart,
			calendarEnd,
			data.termLabel,
		);
		if (!calendar.includes('BEGIN:VEVENT')) {
			exportError = 'No hay horarios dentro de ese rango.';
			return;
		}
		downloadICalendar(calendar);
		calendarModalOpen = false;
		shareFeedback = 'Archivo .ics descargado';
	}
</script>

<svelte:head><title>Horarios | Gestión universitaria</title></svelte:head>

<div class="flex min-h-dvh flex-col lg:h-dvh lg:overflow-hidden">
	<TopBar
		{user}
		{busy}
		{onNavigate}
		{onLogout}
		showShare
		{sharing}
		shareDisabled={summary.selectedCourses === 0}
		onShare={shareBoard}
	/>

	<main
		id="main-content"
		class="mx-auto flex min-h-0 w-full max-w-[2400px] flex-1 flex-col gap-3 overflow-x-clip px-3 py-3 sm:px-4 lg:overflow-hidden xl:px-4"
	>
		{#if loadError}
			<section class="neo-panel grid flex-1 place-items-center p-8 text-center" role="alert">
				<div>
					<h1 class="text-xl font-extrabold text-primary">Los horarios no están disponibles</h1>
					<p class="mt-2 text-sm text-secondary">{loadError}</p>
					<button
						class="mt-5 h-11 rounded-xl bg-accent-strong px-4 text-sm font-bold text-white"
						type="button"
						on:click={load}>Reintentar</button
					>
				</div>
			</section>
		{:else if !data}
			<section class="neo-panel grid flex-1 place-items-center" aria-live="polite">
				<div class="text-center">
					<span
						class="mx-auto block h-8 w-8 animate-spin rounded-full border-2 border-border-strong border-t-accent"
					></span>
					<p class="mt-3 text-sm font-semibold text-secondary">Cargando horarios…</p>
				</div>
			</section>
		{:else}
			<div
				class="neo-control mobile-view-tabs grid grid-cols-2 gap-[3px] p-[3px] lg:hidden"
				role="tablist"
				aria-label="Vista del dashboard"
			>
				<button
					class="min-h-7 rounded-[8px] text-xs font-bold text-secondary"
					class:bg-surface={mobileView === 'courses'}
					class:shadow-card={mobileView === 'courses'}
					class:text-primary={mobileView === 'courses'}
					type="button"
					role="tab"
					aria-selected={mobileView === 'courses'}
					on:click={() => (mobileView = 'courses')}
					>Cursos <span class="text-xs text-muted">{summary.selectedCourses}</span></button
				>
				<button
					class="min-h-7 rounded-[8px] text-xs font-bold text-secondary"
					class:bg-surface={mobileView === 'planner'}
					class:shadow-card={mobileView === 'planner'}
					class:text-primary={mobileView === 'planner'}
					type="button"
					role="tab"
					aria-selected={mobileView === 'planner'}
					on:click={() => (mobileView = 'planner')}
					>Horario {summary.conflictCount > 0 ? `· ${summary.conflictCount}` : ''}</button
				>
			</div>

			<div
				class="grid min-h-0 w-full min-w-0 flex-1 gap-3 lg:grid-cols-[minmax(0,1fr)_minmax(380px,460px)]"
			>
				<div
					class="order-2 min-h-[32rem] min-w-0 flex-1 overflow-auto rounded-[17px] lg:order-1 lg:flex lg:min-h-0"
					class:hidden={mobileView !== 'planner'}
				>
					<WeeklyPlanner
						days={data.days}
						academicHours={data.academicHours}
						{events}
						onOpenEvent={openEvent}
					/>
				</div>
				<div
					class="order-1 min-h-[32rem] min-w-0 lg:order-2 lg:block lg:min-h-0"
					class:hidden={mobileView !== 'courses'}
				>
					<CourseExplorer
						termLabel={data.termLabel}
						{searchQuery}
						{selectedYears}
						{availableYears}
						courses={filteredCourses}
						selectedGroups={$selection}
						{summary}
						{conflictingCourseIds}
						showExports={!!user}
						onSearchChange={(value) => (searchQuery = value)}
						onYearsChange={setSelectedYears}
						onToggleGroup={toggleGroup}
						onClearSelection={clearSelection}
						onPreviewImage={previewImage}
						onExportCalendar={() => {
							if (!user) return;
							exportError = '';
							calendarModalOpen = true;
						}}
						onOpenDetails={(course) => {
							detailCourse = course;
							focusedSessionId = null;
						}}
					/>
				</div>
			</div>
		{/if}
	</main>

	{#if shareFeedback}
		<div
			class="glass-panel fixed right-3 bottom-3 z-30 w-[calc(100%-1.5rem)] max-w-md rounded-2xl p-3"
			role="status"
			aria-live="polite"
		>
			<div class="flex items-center justify-between gap-3">
				<strong class="text-sm text-primary">{shareFeedback}</strong><button
					class="grid h-10 w-10 place-items-center rounded-xl text-secondary hover:bg-surface-muted"
					type="button"
					aria-label="Cerrar aviso"
					on:click={dismissShare}>×</button
				>
			</div>
			{#if shareUrl}<div class="mt-2 flex gap-2">
					<input
						class="neo-control h-11 min-w-0 flex-1 px-3 text-xs text-primary"
						readonly
						value={shareUrl}
						on:focus={(event) => event.currentTarget.select()}
					/><button
						class="neo-button h-11 px-3 text-xs font-bold text-primary"
						type="button"
						on:click={copyShareUrl}>Copiar</button
					>
				</div>{/if}
		</div>
	{/if}

	{#if imagePreviewUrl}<Modal title="Previsualizar" onClose={() => (imagePreviewUrl = '')}>
			<img
				class="w-full rounded-2xl border border-border-subtle bg-surface"
				src={imagePreviewUrl}
				alt="Vista previa del horario semanal"
			/>
			<div class="mt-4 flex justify-end">
				<a
					class="inline-flex min-h-11 items-center rounded-xl bg-accent-strong px-4 text-sm font-bold text-white"
					href={imagePreviewUrl}
					download="2026B.png">Descargar PNG</a
				>
			</div>
		</Modal>{/if}

	{#if calendarModalOpen && user}<Modal
			title="Descargar para Calendar"
			onClose={() => (calendarModalOpen = false)}
		>
			<form class="grid gap-3 sm:grid-cols-2" on:submit|preventDefault={exportCalendar}>
				<label class="flex flex-col gap-1.5"
					><span class="text-[10px] font-extrabold tracking-[0.16em] text-muted uppercase"
						>Desde</span
					><input
						class="neo-control min-h-11 px-3 text-sm text-primary outline-none"
						bind:value={calendarStart}
						required
						type="date"
					/></label
				>
				<label class="flex flex-col gap-1.5"
					><span class="text-[10px] font-extrabold tracking-[0.16em] text-muted uppercase"
						>Hasta</span
					><input
						class="neo-control min-h-11 px-3 text-sm text-primary outline-none"
						bind:value={calendarEnd}
						min={calendarStart}
						required
						type="date"
					/></label
				>
				<p class="text-xs leading-5 text-secondary sm:col-span-2">
					Se descargará un archivo <strong class="text-primary">.ics</strong> con eventos semanales para
					importarlo en Calendar.
				</p>
				{#if exportError}<p
						class="rounded-xl bg-warning-soft px-3 py-2 text-sm font-semibold text-warning sm:col-span-2"
						role="alert"
					>
						{exportError}
					</p>{/if}
				<button
					class="min-h-11 rounded-xl bg-accent-strong px-4 text-sm font-bold text-white sm:col-span-2"
					type="submit">Descargar .ics</button
				>
			</form>
		</Modal>{/if}

	{#if detailCourse}
		<CourseDetailModal
			course={detailCourse}
			courses={allCourses}
			selectedGroups={$selection}
			academicHours={data?.academicHours ?? []}
			{events}
			{focusedSessionId}
			onClose={() => {
				detailCourse = null;
				focusedSessionId = null;
			}}
		/>
	{/if}
</div>

<style>
	.mobile-view-tabs:focus-within {
		border-color: var(--ui-border);
		background: var(--ui-surface-muted);
		box-shadow: var(--ui-shadow-control);
	}

	.mobile-view-tabs button {
		-webkit-tap-highlight-color: transparent;
	}
</style>
