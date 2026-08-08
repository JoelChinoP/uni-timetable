<script lang="ts">
	import { onMount } from 'svelte';

	import TopBar from '../components/TopBar.svelte';
	import CourseExplorer from '../components/CourseExplorer.svelte';
	import WeeklyPlanner from '../components/WeeklyPlanner.svelte';
	import CourseDetailModal from '../components/CourseDetailModal.svelte';
	import { getDashboard, createSharedTimetable } from '../api/planner';
	import { selection, toggleGroup, clearSelection, pruneSelection } from '../stores/selection';
	import {
		buildPlannerEvents,
		buildPlannerSummary,
		getSelectedCourseGroups,
		groupRelatedCourses,
		matchesCourseSearch,
	} from '../utils/planner';
	import type { Course, PlannerConflict, PlannerData, PlannerEvent } from '../types/planner';
	import type { AuthUser } from '../types/auth';

	export let user: AuthUser | null = null;
	export let busy = false;
	export let onNavigate: (path: string) => void;
	export let onLogout: () => void;

	let data: PlannerData | null = null;
	let loadError = '';
	let searchQuery = '';
	let selectedYear: number | null = null;
	let detailCourse: Course | null = null;
	let focusedSessionId: number | null = null;
	let sharing = false;
	let shareFeedback = '';
	let shareUrl = '';
	let mobileView: 'courses' | 'planner' = 'courses';
	let events: PlannerEvent[] = [];
	let conflicts: PlannerConflict[] = [];

	async function load() {
		loadError = '';
		try {
			data = await getDashboard();
			pruneSelection(data.courses);
		} catch {
			loadError = 'No se pudo cargar la oferta de cursos.';
		}
	}

	onMount(() => void load());

	$: allCourses = data?.courses ?? [];
	$: availableYears = [...new Set(allCourses.map((course) => course.academicYear))].sort();
	$: visibleBundles = groupRelatedCourses(allCourses).filter((bundle) => {
		const related = [...(bundle.theory ? [bundle.theory] : []), ...bundle.laboratories];
		return (
			(!selectedYear || related.some((course) => course.academicYear === selectedYear)) &&
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
		shareFeedback = '';
		try {
			const { id } = await createSharedTimetable($selection);
			shareUrl = `${window.location.origin}/s/${id}`;
			try {
				await navigator.clipboard.writeText(shareUrl);
				shareFeedback = 'Enlace copiado';
			} catch {
				shareFeedback = 'Enlace listo para copiar';
			}
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
		class="mx-auto flex min-h-0 w-full max-w-[1920px] flex-1 flex-col gap-3 overflow-x-clip px-3 py-3 sm:px-4 lg:overflow-hidden xl:px-5"
	>
		{#if loadError}
			<section class="neo-panel grid flex-1 place-items-center p-8 text-center" role="alert">
				<div>
					<h1 class="text-xl font-extrabold text-primary">La oferta no está disponible</h1>
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
					<p class="mt-3 text-sm font-semibold text-secondary">Cargando oferta académica…</p>
				</div>
			</section>
		{:else}
			<div
				class="neo-control grid grid-cols-2 gap-1 p-1 lg:hidden"
				role="tablist"
				aria-label="Vista del dashboard"
			>
				<button
					class="min-h-11 rounded-[10px] text-sm font-bold text-secondary"
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
					class="min-h-11 rounded-[10px] text-sm font-bold text-secondary"
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
				class="grid min-h-0 w-full min-w-0 flex-1 gap-3 lg:grid-cols-[minmax(0,1fr)_minmax(330px,370px)]"
			>
				<div
					class="order-2 min-h-[32rem] min-w-0 flex-1 lg:order-1 lg:flex lg:min-h-0"
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
						{selectedYear}
						{availableYears}
						courses={filteredCourses}
						selectedGroups={$selection}
						{summary}
						onSearchChange={(value) => (searchQuery = value)}
						onYearChange={(value) => (selectedYear = value)}
						onToggleGroup={toggleGroup}
						onClearSelection={clearSelection}
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
					on:click={() => (shareFeedback = '')}>×</button
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
