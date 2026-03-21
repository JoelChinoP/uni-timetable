<script lang="ts">
	import { onDestroy, onMount } from 'svelte';

	import CourseDetailModal from './lib/components/CourseDetailModal.svelte';
	import CourseExplorer from './lib/components/CourseExplorer.svelte';
	import HeaderBar from './lib/components/HeaderBar.svelte';
	import PlannerSummaryBar from './lib/components/PlannerSummaryBar.svelte';
	import WeeklyPlanner from './lib/components/WeeklyPlanner.svelte';
	import { loadPlannerDashboard } from './lib/api/planner';
	import { theme, type ThemeMode } from './lib/stores/theme';
	import type { Course, PlannerDashboard, PlannerEvent, PlannerSummary } from './lib/types/planner';
	import {
		buildPlannerEvents,
		buildPlannerSummary,
		getSelectedCourseGroups,
		matchesCourseSearch,
	} from './lib/utils/planner';

	const emptySummary: PlannerSummary = {
		selectedCourses: 0,
		weeklyHours: 0,
		conflictCount: 0,
	};

	let dashboard: PlannerDashboard | null = null;
	let isLoading = true;
	let errorMessage = '';
	let searchQuery = '';
	let activeCourse: Course | null = null;
	let focusedSessionId: number | null = null;
	let themeMode: ThemeMode = 'dark';

	const unsubscribe = theme.subscribe((value) => {
		themeMode = value;
	});

	onDestroy(unsubscribe);

	onMount(async () => {
		try {
			dashboard = await loadPlannerDashboard();
		} catch (error) {
			errorMessage =
				error instanceof Error ? error.message : 'No se pudo cargar el planner.';
		} finally {
			isLoading = false;
		}
	});

	$: courses = dashboard?.courses ?? [];
	$: selectedGroups = dashboard?.selectedGroups ?? {};
	$: filteredCourses = courses.filter((course) => matchesCourseSearch(course, searchQuery));
	$: selectedCourseGroups = getSelectedCourseGroups(courses, selectedGroups);
	$: plannerState = buildPlannerEvents(selectedCourseGroups);
	$: plannerEvents = plannerState.events;
	$: plannerConflicts = plannerState.conflicts;
	$: summary = dashboard
		? buildPlannerSummary(selectedCourseGroups, dashboard.academicHours, plannerConflicts)
		: emptySummary;
	$: activeCourseId = activeCourse?.id ?? null;
	$: activeSelectedGroup =
		activeCourseId == null
			? null
			: (activeCourse?.groups.find(
					(group) => group.id === selectedGroups[String(activeCourseId)],
				) ?? null);

	function handleGroupToggle(courseId: number, groupId: number) {
		if (!dashboard) {
			return;
		}

		const key = String(courseId);
		const currentGroupId = dashboard.selectedGroups[key];

		dashboard = {
			...dashboard,
			selectedGroups: {
				...dashboard.selectedGroups,
				[key]: currentGroupId === groupId ? null : groupId,
			},
		};

		if (activeCourse?.id === courseId) {
			focusedSessionId = null;
		}
	}

	function handleSearchChange(value: string) {
		searchQuery = value;
	}

	function openCourseDetails(course: Course, sessionId: number | null = null) {
		activeCourse = course;
		focusedSessionId = sessionId;
	}

	function openEventDetails(event: PlannerEvent) {
		const course = courses.find(({ id }) => id === event.courseId);
		if (course) {
			openCourseDetails(course, event.sessionId);
		}
	}

	function closeCourseDetails() {
		activeCourse = null;
		focusedSessionId = null;
	}
</script>

{#if isLoading}
	<div class="flex min-h-full items-center justify-center px-4 py-8 sm:px-6">
		<div
			class="w-full max-w-2xl rounded-[32px] border border-border-subtle bg-panel p-8 shadow-panel backdrop-blur-xl"
		>
			<span
				class="inline-flex rounded-full bg-accent-soft px-4 py-1 text-[11px] font-extrabold uppercase tracking-[0.24em] text-accent"
			>
				Planner
			</span>
			<h1 class="mt-5 font-display text-4xl leading-none text-primary sm:text-5xl">
				Preparando tu horario.
			</h1>
			<p class="mt-3 max-w-xl text-sm leading-6 text-secondary sm:text-base">
				Cargando cursos, secciones y bloques semanales para dibujar el tablero.
			</p>
		</div>
	</div>
{:else if errorMessage}
	<div class="flex min-h-full items-center justify-center px-4 py-8 sm:px-6">
		<div
			class="w-full max-w-2xl rounded-[32px] border border-warning/20 bg-panel p-8 shadow-panel backdrop-blur-xl"
		>
			<span
				class="inline-flex rounded-full bg-warning-soft px-4 py-1 text-[11px] font-extrabold uppercase tracking-[0.24em] text-warning"
			>
				Error
			</span>
			<h1 class="mt-5 font-display text-4xl leading-none text-primary sm:text-5xl">
				No pudimos cargar el planner.
			</h1>
			<p class="mt-3 max-w-xl text-sm leading-6 text-secondary sm:text-base">{errorMessage}</p>
		</div>
	</div>
{:else if dashboard}
	<div class="min-h-full">
		<HeaderBar
			navigation={dashboard.navigation}
			user={dashboard.user}
			{themeMode}
			onToggleTheme={() => theme.toggle()}
		/>

		<main class="mx-auto flex w-full max-w-[1880px] flex-1 flex-col px-4 pb-6 pt-4 sm:px-6 xl:px-8">
			<div class="grid min-h-0 flex-1 gap-6 xl:grid-cols-[360px_minmax(0,1fr)]">
				<CourseExplorer
					termLabel={dashboard.termLabel}
					{searchQuery}
					courses={filteredCourses}
					{selectedGroups}
					onSearchChange={handleSearchChange}
					onToggleGroup={handleGroupToggle}
					onOpenDetails={openCourseDetails}
				/>

				<section class="flex min-h-0 flex-col gap-4">
					<WeeklyPlanner
						boardTitle={dashboard.boardTitle}
						boardSubtitle={dashboard.boardSubtitle}
						tabs={dashboard.tabs}
						days={dashboard.days}
						academicHours={dashboard.academicHours}
						events={plannerEvents}
						conflicts={plannerConflicts}
						onOpenEvent={openEventDetails}
					/>

					<PlannerSummaryBar {summary} />
				</section>
			</div>
		</main>

		{#if activeCourse}
			<CourseDetailModal
				course={activeCourse}
				selectedGroup={activeSelectedGroup}
				academicHours={dashboard.academicHours}
				{focusedSessionId}
				events={plannerEvents.filter((event) => event.courseId === activeCourseId)}
				onClose={closeCourseDetails}
			/>
		{/if}
	</div>
{/if}
