<script lang="ts">
	import CourseCard from './CourseCard.svelte';
	import { groupRelatedCourses } from '../utils/planner';
	import type { Course } from '../types/planner';

	export let termLabel: string;
	export let searchQuery = '';
	export let selectedYear: number | null = null;
	export let availableYears: number[] = [];
	export let courses: Course[] = [];
	export let selectedGroups: Record<string, number> = {};
	export let summary: { selectedCourses: number; conflictCount: number };
	export let onSearchChange: (value: string) => void;
	export let onYearChange: (year: number | null) => void;
	export let onToggleGroup: (courseId: number, groupId: number) => void;
	export let onClearSelection: () => void;
	export let onOpenDetails: (course: Course) => void;

	$: bundles = groupRelatedCourses(courses);
</script>

<aside class="neo-panel flex h-full min-h-0 flex-col overflow-hidden p-3">
	<header class="flex items-center justify-between gap-2 px-1">
		<div class="mb-2">
			<h1 class="text-sm font-extrabold text-primary xl:text-[18px]">Cursos</h1>
			<p class="-mt-0.5 text-[11px] font-semibold text-muted">{termLabel}</p>
		</div>
		<div
			class={`rounded-xl px-2.5 py-1.5 text-[10px] font-bold ${summary.conflictCount > 0 ? 'bg-warning-soft text-warning' : 'bg-success-soft text-success'}`}
			role="status"
			aria-live="polite"
		>
			{summary.conflictCount > 0
				? `${summary.conflictCount} conflicto${summary.conflictCount === 1 ? '' : 's'}`
				: 'Sin conflictos'}
		</div>
	</header>

	<div class="grid grid-cols-[minmax(0,1fr)_auto_44px] gap-1">
		<label
			class="neo-control flex min-w-0 items-center gap-1 px-2 text-[16px]"
			aria-label="Buscar cursos"
		>
			<svg
				class="h-5 w-5 shrink-0 stroke-muted"
				viewBox="0 0 24 24"
				fill="none"
				stroke-width="1.9"
				stroke-linecap="round"
				aria-hidden="true"><circle cx="11" cy="11" r="7" /><path d="m16 16 4 4" /></svg
			>
			<input
				class="h-9 min-w-0 flex-1 bg-transparent text-primary outline-none placeholder:text-muted"
				type="search"
				placeholder="Buscar…"
				value={searchQuery}
				on:input={(event) => onSearchChange((event.currentTarget as HTMLInputElement).value)}
			/>
		</label>
		<label class="neo-control grid px-1 text-[16px]" aria-label="Filtrar por año">
			<select
				class="min-w-0 bg-transparent px-1 font-bold text-primary outline-none"
				value={selectedYear ?? ''}
				on:change={(event) =>
					onYearChange(
						(event.currentTarget as HTMLSelectElement).value
							? Number((event.currentTarget as HTMLSelectElement).value)
							: null,
					)}
			>
				<option value="">#</option>
				{#each availableYears as year (year)}<option value={year}>{year}°</option>{/each}
			</select>
		</label>
		<button
			class="neo-button grid h-10 w-10 place-items-center text-secondary disabled:opacity-40"
			type="button"
			title="Limpiar selección"
			aria-label="Limpiar selección"
			disabled={summary.selectedCourses === 0}
			on:click={onClearSelection}
		>
			<svg
				class="h-6 w-6 stroke-current"
				viewBox="0 0 24 24"
				fill="none"
				stroke-width="1.9"
				stroke-linecap="round"
				aria-hidden="true"><path d="M4 7h16M9 7V5h6v2M7 7l1 13h8l1-13" /></svg
			>
		</button>
	</div>

	<div class="mt-0.75 flex items-center justify-between px-4 text-[11px] text-secondary">
		<span><strong class="text-primary">{summary.selectedCourses}</strong> seleccionados</span>
		<span>{bundles.length} cursos</span>
	</div>

	<div class="mt-1 min-h-0 flex-1 overflow-y-auto pr-1">
		{#if bundles.length === 0}
			<div class="rounded-2xl border border-dashed border-border-strong p-4">
				<h2 class="text-sm font-bold text-primary">Sin resultados</h2>
				<p class="mt-1 text-xs leading-5 text-secondary">Prueba otra búsqueda o cambia el año.</p>
			</div>
		{:else}
			<div class="space-y-2">
				{#each bundles as bundle (bundle.key)}
					<CourseCard {bundle} {selectedGroups} {onToggleGroup} {onOpenDetails} />
				{/each}
			</div>
		{/if}
	</div>
</aside>
