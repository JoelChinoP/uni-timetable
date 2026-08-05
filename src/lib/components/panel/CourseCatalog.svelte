<script lang="ts">
	import type { AcademicHour, Course, CourseGroup } from '../../types/planner';
	import { formatTimeRange, getDayLabel, getModeLabel } from '../../utils/planner';

	export let courses: Course[] = [];
	export let academicHours: AcademicHour[] = [];
	export let onAddGroup: (course: Course) => void;
	export let onDeleteGroup: (group: CourseGroup, course: Course) => void;
	export let onDeleteCourse: (course: Course) => void;

	let search = '';
	let selectedId: number | null = null;

	$: visible = courses.filter((course) => {
		const query = search.trim().toLowerCase();
		if (!query) {
			return true;
		}
		return (
			course.name.toLowerCase().includes(query) ||
			course.abbreviation.toLowerCase().includes(query) ||
			course.code.toLowerCase().includes(query)
		);
	});

	$: selected = visible.find((course) => course.id === selectedId) ?? visible[0] ?? null;
</script>

<div
	class="grid divide-y divide-grid lg:grid-cols-[minmax(0,270px)_minmax(0,1fr)] lg:divide-x lg:divide-y-0"
>
	<div class="flex max-h-[42vh] min-h-0 flex-col lg:h-[62vh] lg:max-h-none">
		<label
			class="flex items-center border-b border-border-subtle px-3 py-2.5"
			aria-label="Buscar cursos"
		>
			<svg
				class="h-3.5 w-3.5 shrink-0 stroke-muted"
				viewBox="0 0 24 24"
				fill="none"
				stroke-width="2"
				stroke-linecap="round"
				stroke-linejoin="round"
				aria-hidden="true"
			>
				<path d="m21 21-4.35-4.35m1.85-5.15a7 7 0 1 1-14 0a7 7 0 0 1 14 0Z" />
			</svg>
			<input
				class="min-w-0 flex-1 bg-transparent px-2 text-xs text-primary outline-none placeholder:text-muted"
				type="search"
				placeholder="Buscar curso…"
				bind:value={search}
			/>
		</label>

		<div class="min-h-0 flex-1 overflow-y-auto py-1">
			{#if visible.length === 0}
				<p class="px-3 py-4 text-xs text-secondary">Sin cursos que coincidan.</p>
			{/if}
			{#each visible as course (course.id)}
				<button
					class={`flex w-full items-center gap-2 px-3 py-2 text-left transition ${
						selected?.id === course.id ? 'bg-accent-soft' : 'hover:bg-surface-muted'
					}`}
					type="button"
					on:click={() => (selectedId = course.id)}
				>
					<span
						class="h-2 w-2 shrink-0 rounded-full"
						style={`background:${course.color};`}
						aria-hidden="true"
					></span>
					<span class="min-w-0 flex-1">
						<span class="block truncate text-xs font-semibold text-primary">{course.name}</span>
						<span class="block text-[10px] font-semibold tracking-wide text-muted uppercase">
							{course.abbreviation} · Año {course.academicYear}
							{course.type === 'LABORATORY' ? ' · Lab' : ''}
						</span>
					</span>
					<span class="text-[10px] font-bold text-muted">{course.groups.length}</span>
				</button>
			{/each}
		</div>
	</div>

	<div class="min-h-0 lg:h-[62vh] lg:overflow-y-auto">
		{#if !selected}
			<div class="grid h-full min-h-40 place-items-center p-6 text-sm text-secondary">
				Selecciona un curso para ver sus grupos.
			</div>
		{:else}
			<div class="border-b border-border-subtle px-4 py-3">
				<div class="flex flex-wrap items-center justify-between gap-2">
					<div class="min-w-0">
						<h3 class="truncate text-sm font-bold text-primary">{selected.name}</h3>
						<p class="mt-0.5 text-[10px] font-semibold tracking-wide text-muted uppercase">
							{selected.abbreviation} · Año {selected.academicYear} · {getModeLabel(selected.type)}
							{selected.teacher ? ` · ${selected.teacher.fullName}` : ''}
						</p>
					</div>
					<div class="flex items-center gap-2">
						<button
							class="rounded-[10px] bg-accent-strong px-3 py-1.5 text-xs font-bold text-white transition hover:bg-accent"
							type="button"
							on:click={() => selected && onAddGroup(selected)}
						>
							+ Grupo
						</button>
						<button
							class="grid h-7 w-7 place-items-center rounded-[10px] text-warning transition hover:bg-warning-soft"
							type="button"
							aria-label={`Eliminar ${selected.name}`}
							on:click={() => selected && onDeleteCourse(selected)}
						>
							<svg
								class="h-3.5 w-3.5 stroke-current"
								viewBox="0 0 24 24"
								fill="none"
								stroke-width="2"
								stroke-linecap="round"
								aria-hidden="true"
							>
								<path d="M4 7h16M9 7V5a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v2m-9 0 1 13h8l1-13" />
							</svg>
						</button>
					</div>
				</div>
			</div>

			<div class="space-y-2 p-4">
				{#if selected.groups.length === 0}
					<p
						class="rounded-[14px] border border-dashed border-border-strong px-4 py-3 text-xs text-secondary"
					>
						Sin grupos registrados. Usa "+ Grupo" para crear el primero.
					</p>
				{/if}
				{#each selected.groups as group (group.id)}
					<article class="rounded-[14px] border border-border-subtle bg-surface p-3">
						<div class="flex items-center justify-between gap-2">
							<div class="flex items-center gap-2">
								<span
									class="rounded-lg px-2.5 py-1 text-xs font-extrabold text-white"
									style={`background:${selected.color};`}
								>
									Grupo {group.name}
								</span>
								{#if group.classroomLabel}
									<span class="text-xs font-semibold text-secondary">
										{group.classroomLabel}
									</span>
								{/if}
							</div>
							<button
								class="rounded-lg px-2 py-1 text-[11px] font-bold text-warning transition hover:bg-warning-soft"
								type="button"
								on:click={() => selected && onDeleteGroup(group, selected)}
							>
								Eliminar
							</button>
						</div>
						<div class="mt-2 flex flex-wrap gap-1.5">
							{#each group.sessions as session (session.id)}
								<span
									class="rounded-[10px] bg-surface-muted px-2.5 py-1 text-[11px] text-secondary"
								>
									<strong class="text-primary">{getDayLabel(session.day).slice(0, 3)}</strong>
									{formatTimeRange(session.startHourAcademic, session.durationHours, academicHours)}
								</span>
							{/each}
						</div>
					</article>
				{/each}
			</div>
		{/if}
	</div>
</div>
