<script lang="ts">
	import type { Course } from '../types/planner';
	import CourseCard from './CourseCard.svelte';

	export let termLabel: string;
	export let searchQuery = '';
	export let yearsSelected: Record<number, boolean> = {};
	export let availableYears: number[] = [];
	export let courses: Course[] = [];
	export let selectedGroups: Record<string, number> = {};
	export let summary: { selectedCourses: number; conflictCount: number };
	export let sharing = false;
	export let shareFeedback = '';
	export let onSearchChange: (value: string) => void;
	export let onToggleYear: (year: number) => void;
	export let onToggleGroup: (courseId: number, groupId: number) => void;
	export let onClearSelection: () => void;
	export let onOpenDetails: (course: Course) => void;
	export let onShare: () => void;
</script>

<aside
	class="flex h-full min-h-0 flex-col overflow-hidden rounded-[20px] border border-border-subtle bg-panel px-3 py-3 shadow-card backdrop-blur-xl"
>
	<div class="flex items-center justify-between gap-2">
		<h1 class="font-display text-base font-bold text-primary">Horarios</h1>
		<div class="flex items-center gap-2">
			<span class="text-[10px] font-medium text-muted">{termLabel}</span>
			<button
				class="grid h-7 w-7 place-items-center rounded-lg border border-border-subtle bg-surface text-secondary shadow-card transition hover:text-accent disabled:cursor-not-allowed disabled:opacity-50"
				type="button"
				title={summary.selectedCourses === 0 ? 'Selecciona al menos un grupo' : 'Compartir horario'}
				aria-label="Compartir horario"
				disabled={sharing || summary.selectedCourses === 0}
				on:click={onShare}
			>
				<svg
					class="h-3.5 w-3.5 stroke-current"
					viewBox="0 0 24 24"
					fill="none"
					stroke-width="2"
					stroke-linecap="round"
					stroke-linejoin="round"
					aria-hidden="true"
				>
					<circle cx="18" cy="5" r="3" />
					<circle cx="6" cy="12" r="3" />
					<circle cx="18" cy="19" r="3" />
					<path d="m8.6 13.5 6.8 4M15.4 6.5l-6.8 4" />
				</svg>
			</button>
		</div>
	</div>

	<div class="mt-1 flex items-center justify-between gap-2">
		<p class="text-[11px] text-secondary">
			<strong class="font-bold text-primary">{summary.selectedCourses}</strong> cursos ·
			<strong class={`font-bold ${summary.conflictCount > 0 ? 'text-warning' : 'text-primary'}`}
				>{summary.conflictCount}</strong
			>
			cruces
		</p>
		{#if summary.selectedCourses > 0}
			<button
				class="text-[10px] font-bold text-warning transition hover:underline"
				type="button"
				on:click={onClearSelection}
			>
				Limpiar
			</button>
		{/if}
	</div>
	{#if shareFeedback}
		<p class="mt-1 text-[10px] text-secondary">{shareFeedback}</p>
	{/if}

	<label
		class="mt-2.5 flex items-center gap-2 rounded-[12px] border border-border-subtle bg-surface px-2.5 py-2 shadow-card transition duration-200 focus-within:border-accent/35 focus-within:ring-2 focus-within:ring-accent-soft"
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
			class="min-w-0 flex-1 bg-transparent text-xs text-primary outline-none placeholder:text-muted"
			type="search"
			placeholder="Buscar por nombre o sigla…"
			value={searchQuery}
			on:input={(event) => onSearchChange((event.currentTarget as HTMLInputElement).value)}
		/>
	</label>

	<div class="mt-2 flex items-center gap-1" title="Filtra por año; sin selección muestra todos">
		{#each availableYears as year (year)}
			<button
				class={`grid h-6 w-7 place-items-center rounded-md text-[11px] font-bold transition duration-150 ${
					yearsSelected[year]
						? 'bg-accent-strong text-white'
						: 'bg-surface-muted text-secondary hover:text-primary'
				}`}
				type="button"
				aria-pressed={!!yearsSelected[year]}
				on:click={() => onToggleYear(year)}
			>
				{year}
			</button>
		{/each}
	</div>

	<div class="mt-2.5 min-h-0 flex-1 overflow-y-auto pr-1">
		<div class="space-y-1.5">
			{#if courses.length === 0}
				<div class="rounded-[14px] border border-dashed border-border-strong bg-surface p-3.5">
					<h3 class="text-sm font-semibold text-primary">Sin resultados</h3>
					<p class="mt-1 text-[11px] leading-4 text-secondary">
						Prueba otra búsqueda o quita el filtro de año.
					</p>
				</div>
			{:else}
				{#each courses as course (course.id)}
					<CourseCard
						{course}
						selectedGroupId={selectedGroups[String(course.id)] ?? null}
						{onToggleGroup}
						{onOpenDetails}
					/>
				{/each}
			{/if}
		</div>
	</div>
</aside>
