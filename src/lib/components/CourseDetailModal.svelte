<script lang="ts">
	import { fade, scale } from 'svelte/transition';
	import { trapFocus } from '../utils/focus';
	import type { AcademicHour, Course, PlannerEvent, SessionMode } from '../types/planner';
	import { formatTimeRange, getDayLabel } from '../utils/planner';

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
		class="glass-panel flex max-h-[94dvh] w-full max-w-4xl flex-col overflow-hidden rounded-[24px]"
		role="dialog"
		aria-modal="true"
		aria-labelledby="course-detail-title"
		tabindex="-1"
		use:trapFocus
		transition:scale={{ start: 0.97, duration: 170 }}
	>
		<header
			class="flex items-start justify-between gap-4 border-b border-border-subtle px-4 py-4 sm:px-6"
		>
			<div class="min-w-0">
				<div class="flex items-center gap-2">
					<span
						class="h-2.5 w-2.5 rounded-full"
						style={`background:${primary.color};`}
						aria-hidden="true"
					></span><span class="text-[10px] font-extrabold tracking-[0.18em] text-accent uppercase"
						>{primary.abbreviation.replace(/-L$/, '')} · {primary.academicYear}° año</span
					>
				</div>
				<h2
					id="course-detail-title"
					class="mt-2 font-display text-2xl leading-tight font-extrabold text-primary sm:text-3xl"
				>
					{primary.name}
				</h2>
				<p class="mt-2 text-xs text-secondary sm:text-sm">
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
				><svg
					class="h-5 w-5 stroke-current"
					viewBox="0 0 24 24"
					fill="none"
					stroke-width="1.9"
					stroke-linecap="round"
					aria-hidden="true"><path d="m6 6 12 12M18 6 6 18" /></svg
				></button
			>
		</header>

		<div class="border-b border-border-subtle px-4 py-3 sm:px-6">
			<div
				class="neo-control grid max-w-sm grid-cols-2 gap-1 p-1"
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

		<div class="min-h-0 flex-1 overflow-y-auto p-4 sm:p-6" role="tabpanel">
			{#if primary.summary}<p class="mb-4 max-w-3xl text-sm leading-6 text-secondary">
					{primary.summary}
				</p>{/if}
			<div class="grid gap-4 xl:grid-cols-2">
				{#each visibleCourses as visibleCourse (visibleCourse.id)}
					<section class="neo-card p-3 sm:p-4">
						<div class="flex items-center justify-between gap-3">
							<div>
								<p class="text-[9px] font-extrabold tracking-[0.18em] text-muted uppercase">
									{visibleCourse.type === 'THEORY' ? 'Teoría' : visibleCourse.abbreviation}
								</p>
								<h3 class="mt-1 text-base font-extrabold text-primary">Grupos y horarios</h3>
							</div>
							<span
								class="rounded-lg bg-surface-muted px-2 py-1 text-[10px] font-bold text-secondary"
								>{visibleCourse.groups.length} grupos</span
							>
						</div>
						{#if visibleCourse.groups.length === 0}
							<p
								class="mt-3 rounded-xl border border-dashed border-border-strong p-3 text-xs text-secondary"
							>
								No hay grupos registrados.
							</p>
						{:else}
							<div class="mt-3 space-y-2">
								{#each visibleCourse.groups as group (group.id)}
									<article
										class={`rounded-xl border p-3 ${selectedGroups[String(visibleCourse.id)] === group.id ? 'border-accent bg-accent-soft' : 'border-border-subtle bg-surface-muted/70'}`}
									>
										<div class="flex items-center justify-between gap-2">
											<strong class="text-sm text-primary">Grupo {group.name}</strong
											>{#if selectedGroups[String(visibleCourse.id)] === group.id}<span
													class="text-[10px] font-bold text-accent">Seleccionado</span
												>{/if}
										</div>
										<div class="mt-2 space-y-1.5">
											{#each group.sessions as session (session.id)}
												<div
													class={`flex flex-wrap items-center justify-between gap-2 rounded-lg px-2.5 py-2 text-xs ${session.id === focusedSessionId ? 'bg-accent-soft ring-2 ring-accent/30' : conflictEventIds.has(session.id) ? 'bg-warning-soft' : 'bg-surface'}`}
												>
													<span class="font-semibold text-primary"
														>{getDayLabel(session.day)} · {formatTimeRange(
															session.startHourAcademic,
															session.durationHours,
															academicHours,
														)}</span
													><span class="text-secondary">{session.classroomLabel || 'Sin aula'}</span
													>{#if conflictEventIds.has(session.id)}<span
															class="font-bold text-warning">Cruce</span
														>{/if}
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
