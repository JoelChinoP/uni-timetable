<script lang="ts">
	import type { AcademicHour, Course } from '../types/planner';
	import CourseCard from './CourseCard.svelte';

	export let termLabel: string;
	export let searchQuery = '';
	export let courses: Course[] = [];
	export let selectedGroups: Record<string, number | null> = {};
	export let academicHours: AcademicHour[] = [];
	export let onSearchChange: (value: string) => void;
	export let onToggleGroup: (courseId: number, groupId: number) => void;
	export let onOpenDetails: (course: Course) => void;
</script>

<aside
	class="flex h-full min-h-0 flex-col overflow-hidden rounded-[24px] border border-border-subtle bg-panel px-3 py-3 shadow-card backdrop-blur-xl"
>
	<div class="border-b border-border-subtle pb-3">
		<div class="flex items-center justify-between gap-3">
			<div>
				<p class="text-[9px] font-extrabold tracking-[0.26em] text-accent uppercase">Cursos</p>
				<h2 class="mt-1 text-sm font-semibold text-primary">Seleccion de horario</h2>
			</div>
			<span class="text-[11px] font-medium text-muted">{termLabel}</span>
		</div>

		<label
			class="mt-3 flex items-center gap-3 rounded-[16px] border border-border-subtle bg-surface px-3 py-2.5 shadow-card transition duration-200 focus-within:border-accent/35 focus-within:ring-2 focus-within:ring-accent-soft"
			aria-label="Buscar cursos"
		>
			<svg
				class="h-4 w-4 shrink-0 stroke-muted"
				viewBox="0 0 24 24"
				fill="none"
				stroke-width="1.8"
				stroke-linecap="round"
				stroke-linejoin="round"
				aria-hidden="true"
			>
				<path d="m21 21-4.35-4.35m1.85-5.15a7 7 0 1 1-14 0a7 7 0 0 1 14 0Z" />
			</svg>
			<input
				class="min-w-0 flex-1 bg-transparent text-sm text-primary outline-none placeholder:text-muted"
				type="search"
				placeholder="Buscar cursos..."
				value={searchQuery}
				on:input={(event) => onSearchChange((event.currentTarget as HTMLInputElement).value)}
			/>
		</label>
	</div>

	<div class="mt-3 min-h-0 flex-1 overflow-y-auto pr-1">
		<div class="space-y-2">
			{#if courses.length === 0}
				<div
					class="rounded-[18px] border border-dashed border-border-strong bg-surface p-4 shadow-card"
				>
					<h3 class="text-sm font-semibold text-primary">Sin resultados</h3>
					<p class="mt-1 text-xs leading-5 text-secondary">
						Prueba otra busqueda para volver a mostrar cursos.
					</p>
				</div>
			{:else}
				{#each courses as course (course.id)}
					<CourseCard
						{course}
						{academicHours}
						selectedGroupId={selectedGroups[String(course.id)] ?? null}
						{onToggleGroup}
						{onOpenDetails}
					/>
				{/each}
			{/if}
		</div>
	</div>
</aside>
