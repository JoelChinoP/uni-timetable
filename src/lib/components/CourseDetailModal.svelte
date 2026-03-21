<script lang="ts">
	import { fade, scale } from 'svelte/transition';

	import type { AcademicHour, Course, CourseGroup, PlannerEvent } from '../types/planner';
	import { formatTimeRange, getDayLabel, getModeLabel } from '../utils/planner';

	export let course: Course;
	export let selectedGroup: CourseGroup | null;
	export let academicHours: AcademicHour[] = [];
	export let events: PlannerEvent[] = [];
	export let focusedSessionId: number | null = null;
	export let onClose: () => void;

	$: conflictEventIds = new Set(
		events.filter((event) => event.conflictIds.length > 0).map(({ sessionId }) => sessionId),
	);

	function handleBackdropClick(event: MouseEvent) {
		if (event.target === event.currentTarget) {
			onClose();
		}
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape') {
			onClose();
		}
	}
</script>

<svelte:window on:keydown={handleKeydown} />

<div
	class="fixed inset-0 z-40 grid place-items-center p-4 backdrop-blur-md sm:p-6"
	role="presentation"
	style="background:var(--ui-overlay);"
	on:click={handleBackdropClick}
	transition:fade
>
	<div
		class="flex max-h-[92vh] w-full max-w-3xl flex-col overflow-hidden rounded-[32px] border border-border-subtle bg-panel-strong shadow-panel"
		role="dialog"
		aria-modal="true"
		aria-labelledby="course-detail-title"
		transition:scale={{ start: 0.96, duration: 160 }}
	>
		<header class="border-b border-border-subtle px-5 py-5 sm:px-6">
			<div class="flex items-start justify-between gap-4">
				<div class="min-w-0">
					<div class="flex flex-wrap items-center gap-2">
						<span class="h-3 w-3 rounded-full" style={`background:${course.color};`}></span>
						<span
							class="text-[10px] font-extrabold uppercase tracking-[0.24em]"
							style={`color:${course.color};`}
						>
							{course.code}
						</span>

						{#if selectedGroup}
							<span
								class="rounded-full bg-surface-muted px-3 py-1 text-[10px] font-extrabold uppercase tracking-[0.22em] text-secondary"
							>
								Grupo {selectedGroup.name}
							</span>
						{/if}
					</div>

					<h2
						id="course-detail-title"
						class="mt-3 font-display text-3xl leading-none text-primary sm:text-4xl"
					>
						{course.name}
					</h2>
					<p class="mt-3 max-w-2xl text-sm leading-6 text-secondary">{course.summary}</p>
				</div>

				<button
					class="grid h-11 w-11 shrink-0 place-items-center rounded-full border border-border-subtle bg-surface text-secondary shadow-card transition duration-200 hover:-translate-y-0.5 hover:border-accent/30 hover:text-primary"
					type="button"
					aria-label="Cerrar detalles del curso"
					on:click={onClose}
				>
					<svg
						class="h-5 w-5 stroke-current"
						viewBox="0 0 24 24"
						fill="none"
						stroke-width="1.9"
						stroke-linecap="round"
						stroke-linejoin="round"
						aria-hidden="true"
					>
						<path d="m6 6 12 12M18 6 6 18" />
					</svg>
				</button>
			</div>
		</header>

		<div class="flex flex-wrap gap-2 border-b border-border-subtle px-5 py-4 sm:px-6">
			<span class="rounded-full bg-surface-muted px-3 py-1.5 text-sm text-secondary">
				{course.teacher.fullName}
			</span>
			<span class="rounded-full bg-surface-muted px-3 py-1.5 text-sm text-secondary">
				{course.credits} creditos
			</span>
			<span class="rounded-full bg-surface-muted px-3 py-1.5 text-sm text-secondary">
				Ano {course.academicYear}
			</span>
			{#if selectedGroup}
				<span class="rounded-full bg-accent-soft px-3 py-1.5 text-sm font-semibold text-accent">
					Seccion activa
				</span>
			{/if}
		</div>

		<div class="min-h-0 flex-1 overflow-y-auto px-5 pb-5 pt-4 sm:px-6">
			{#if selectedGroup}
				<section class="rounded-[28px] border border-border-subtle bg-surface p-4 shadow-card">
					<div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
						<div>
							<p class="text-[10px] font-extrabold uppercase tracking-[0.22em] text-muted">
								Seccion elegida
							</p>
							<h3 class="mt-2 text-xl font-bold text-primary">Horario del grupo {selectedGroup.name}</h3>
						</div>

						<span
							class="rounded-full bg-surface-muted px-3 py-1.5 text-xs font-bold uppercase tracking-[0.18em] text-secondary"
						>
							{selectedGroup.sessions.length} bloques
						</span>
					</div>

					<div class="mt-4 space-y-3">
						{#each selectedGroup.sessions as session (session.id)}
							<article
								class={`flex flex-col gap-3 rounded-[24px] border p-4 sm:flex-row sm:items-start sm:justify-between ${
									session.id === focusedSessionId
										? 'border-accent bg-accent-soft ring-4 ring-accent-soft'
										: conflictEventIds.has(session.id)
											? 'border-warning/25 bg-warning-soft'
											: 'border-border-subtle bg-surface-muted'
								}`}
							>
								<div>
									<strong class="block text-base text-primary">{session.title}</strong>
									<p class="mt-2 text-sm leading-6 text-secondary">
										{getDayLabel(session.day)} - {formatTimeRange(
											session.startHourAcademic,
											session.durationHours,
											academicHours,
										)}
									</p>
								</div>

								<div class="flex flex-wrap gap-2 sm:justify-end">
									<span
										class="rounded-full bg-surface px-3 py-1.5 text-xs font-semibold text-secondary"
									>
										{getModeLabel(session.mode)}
									</span>
									<span
										class="rounded-full bg-surface px-3 py-1.5 text-xs font-semibold text-secondary"
									>
										{session.classroomLabel}
									</span>
									{#if conflictEventIds.has(session.id)}
										<span
											class="rounded-full bg-warning/10 px-3 py-1.5 text-xs font-semibold text-warning"
										>
											Cruce
										</span>
									{/if}
								</div>
							</article>
						{/each}
					</div>
				</section>
			{:else}
				<section
					class="rounded-[28px] border border-dashed border-border-strong bg-surface p-5 shadow-card"
				>
					<h3 class="text-xl font-bold text-primary">Todavia no hay una seccion seleccionada</h3>
					<p class="mt-3 max-w-2xl text-sm leading-6 text-secondary">
						Elige un grupo en el panel izquierdo para ver este curso sobre el tablero semanal.
					</p>
				</section>
			{/if}
		</div>
	</div>
</div>
