<script lang="ts">
	import { fade, scale } from 'svelte/transition';
	import { trapFocus } from '../utils/focus';
	import type { AcademicHour, Course, PlannerEvent, SessionMode } from '../types/planner';
	import {
		formatTimeRange,
		getCourseDisplayCode,
		getCourseDisplayName,
		getDayLabel,
	} from '../utils/planner';

	export let course: Course;
	export let courses: Course[] = [];
	export let selectedGroups: Record<string, number> = {};
	export let academicHours: AcademicHour[] = [];
	export let events: PlannerEvent[] = [];
	export let focusedSessionId: number | null = null;
	export let onClose: () => void;

	let activeTab: SessionMode = course.type;
	$: theory =
		course.type === 'THEORY'
			? course
			: (courses.find(({ id }) => id === course.theoryCourseId) ?? null);
	$: laboratories = courses.filter(
		(item) => item.type === 'LABORATORY' && item.theoryCourseId === theory?.id,
	);
	$: primary = theory ?? course;
	$: visibleCourses =
		activeTab === 'THEORY'
			? theory
				? [theory]
				: []
			: laboratories.length > 0
				? laboratories
				: course.type === 'LABORATORY'
					? [course]
					: [];
	$: modalTitle =
		activeTab === 'LABORATORY' && visibleCourses.length > 1
			? `Laboratorios · ${getCourseDisplayName(primary)}`
			: getCourseDisplayName(visibleCourses[0] ?? primary);
	$: conflictEventIds = new Set(
		events.filter((event) => event.conflictIds.length > 0).map(({ sessionId }) => sessionId),
	);

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape') onClose();
	}
</script>

<svelte:window on:keydown={handleKeydown} />

<div
	class="fixed inset-0 z-40 grid place-items-center bg-[var(--ui-overlay)] p-3 backdrop-blur-sm sm:p-6"
	role="presentation"
	on:click={(event) => event.target === event.currentTarget && onClose()}
	transition:fade
>
	<div
		class="glass-panel flex max-h-[94dvh] w-full max-w-3xl flex-col overflow-hidden rounded-[24px]"
		role="dialog"
		aria-modal="true"
		aria-labelledby="course-detail-title"
		tabindex="-1"
		use:trapFocus
		transition:scale={{ start: 0.97, duration: 170 }}
	>
		<header class="flex items-start gap-3 px-4 py-4 sm:px-5">
			<span
				class="mt-1 h-10 w-1.5 shrink-0 rounded-full"
				style={`background:${primary.color};`}
				aria-hidden="true"
			></span>
			<div class="min-w-0 flex-1">
				<p class="text-[10px] font-extrabold tracking-[0.16em] text-accent uppercase">
					{activeTab === 'THEORY'
						? getCourseDisplayCode(primary)
						: `${visibleCourses.length} laboratorio${visibleCourses.length === 1 ? '' : 's'}`} · {primary.academicYear}°
					año
				</p>
				<h2
					id="course-detail-title"
					class="mt-1 truncate font-display text-xl font-extrabold text-primary sm:text-2xl"
					title={modalTitle}
				>
					{modalTitle}
				</h2>
				<p class="mt-1 text-xs text-secondary">
					{primary.teacher?.fullName ?? 'Docente por asignar'}{primary.credits
						? ` · ${primary.credits} créditos`
						: ''}
				</p>
			</div>
			<button
				class="neo-button grid h-11 w-11 shrink-0 place-items-center text-secondary"
				type="button"
				aria-label="Cerrar detalles"
				on:click={onClose}
			>
				<svg
					class="h-5 w-5 stroke-current"
					viewBox="0 0 24 24"
					fill="none"
					stroke-width="1.9"
					stroke-linecap="round"
					aria-hidden="true"><path d="m6 6 12 12M18 6 6 18" /></svg
				>
			</button>
		</header>

		<div class="px-4 pb-3 sm:px-5">
			{#if primary.summary}<p class="neo-control px-3 py-2.5 text-xs leading-5 text-secondary">
					{primary.summary}
				</p>{/if}
			<div
				class="neo-control mt-3 grid grid-cols-2 gap-1 p-1"
				role="tablist"
				aria-label="Modalidad del curso"
			>
				<button
					class="min-h-10 rounded-[9px] text-xs font-bold text-secondary"
					class:bg-surface={activeTab === 'THEORY'}
					class:shadow-card={activeTab === 'THEORY'}
					class:text-primary={activeTab === 'THEORY'}
					type="button"
					role="tab"
					aria-selected={activeTab === 'THEORY'}
					disabled={!theory}
					on:click={() => (activeTab = 'THEORY')}>Teoría</button
				>
				<button
					class="min-h-10 rounded-[9px] text-xs font-bold text-secondary disabled:opacity-40"
					class:bg-surface={activeTab === 'LABORATORY'}
					class:shadow-card={activeTab === 'LABORATORY'}
					class:text-primary={activeTab === 'LABORATORY'}
					type="button"
					role="tab"
					aria-selected={activeTab === 'LABORATORY'}
					disabled={laboratories.length === 0 && course.type !== 'LABORATORY'}
					on:click={() => (activeTab = 'LABORATORY')}
					>Laboratorio{laboratories.length > 1 ? `s (${laboratories.length})` : ''}</button
				>
			</div>
		</div>

		<div
			class="min-h-0 flex-1 overflow-y-auto border-t border-border-subtle px-4 py-4 sm:px-5"
			role="tabpanel"
		>
			<div class="space-y-4">
				{#each visibleCourses as visibleCourse (visibleCourse.id)}
					<section>
						<div class="flex flex-wrap items-center justify-between gap-2 px-1">
							<div class="min-w-0">
								<h3 class="truncate text-sm font-extrabold text-primary">
									{getCourseDisplayName(visibleCourse)}
								</h3>
								<p class="mt-0.5 text-[10px] font-semibold text-secondary">
									{getCourseDisplayCode(visibleCourse)} · {visibleCourse.teacher?.fullName ??
										'Docente por asignar'}
								</p>
							</div>
							<span
								class="rounded-lg bg-surface-muted px-2 py-1 text-[10px] font-bold text-secondary"
								>{visibleCourse.groups.length} grupos</span
							>
						</div>

						{#if visibleCourse.groups.length === 0}
							<p class="neo-control mt-2 p-3 text-xs text-secondary">No hay grupos registrados.</p>
						{:else}
							<div class="mt-2 space-y-2">
								{#each visibleCourse.groups as group (group.id)}
									<article
										class={`neo-control p-2.5 ${selectedGroups[String(visibleCourse.id)] === group.id ? 'ring-1 ring-accent/55' : ''}`}
									>
										<div class="flex items-center justify-between gap-2 px-1">
											<strong class="text-xs text-primary">Grupo {group.name}</strong>
											{#if selectedGroups[String(visibleCourse.id)] === group.id}<span
													class="inline-flex items-center gap-1 text-[10px] font-bold text-accent"
													><svg
														class="h-3.5 w-3.5 stroke-current"
														viewBox="0 0 24 24"
														fill="none"
														stroke-width="2.5"
														stroke-linecap="round"
														stroke-linejoin="round"
														aria-hidden="true"><path d="m5 12 4 4L19 6" /></svg
													>Elegido</span
												>{/if}
										</div>
										<div class="mt-2 grid gap-1.5 sm:grid-cols-2">
											{#each group.sessions as session (session.id)}
												<div
													class={`rounded-lg bg-surface px-2.5 py-2 text-xs shadow-card ${session.id === focusedSessionId ? 'ring-2 ring-accent/40' : conflictEventIds.has(session.id) ? 'ring-2 ring-warning/35' : ''}`}
												>
													<div class="flex items-center justify-between gap-2">
														<strong class="text-primary">{getDayLabel(session.day)}</strong
														>{#if conflictEventIds.has(session.id)}<span
																class="font-bold text-warning">Cruce</span
															>{/if}
													</div>
													<p class="mt-1 font-semibold text-secondary">
														{formatTimeRange(
															session.startHourAcademic,
															session.durationHours,
															academicHours,
														)}
													</p>
													<p class="mt-0.5 truncate text-[10px] text-muted">
														{session.classroomLabel || 'Sin aula'}
													</p>
												</div>
											{/each}
										</div>
									</article>
								{/each}
							</div>
						{/if}
					</section>
				{/each}
			</div>
		</div>
	</div>
</div>
