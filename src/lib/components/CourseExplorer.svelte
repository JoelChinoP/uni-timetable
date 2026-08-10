<script lang="ts">
	import CourseCard from './CourseCard.svelte';
	import { getAcademicYearLabel, groupRelatedCourses } from '../utils/planner';
	import type { Course } from '../types/planner';

	export let termLabel: string;
	export let searchQuery = '';
	export let selectedYears: number[] = [];
	export let availableYears: number[] = [];
	export let courses: Course[] = [];
	export let selectedGroups: Record<string, number> = {};
	export let summary: { selectedCourses: number; conflictCount: number };
	export let onSearchChange: (value: string) => void;
	export let onYearsChange: (years: number[]) => void;
	export let onToggleGroup: (courseId: number, groupId: number) => void;
	export let onClearSelection: () => void;
	export let onOpenDetails: (course: Course) => void;
	export let conflictingCourseIds: Set<number> = new Set();
	export let showExports = false;
	export let onPreviewImage: () => void = () => {};
	export let onExportCalendar: () => void = () => {};
	let yearMenu: HTMLDetailsElement;
	let yearMenuButton: HTMLElement;

	$: bundles = groupRelatedCourses(courses);
	$: yearLabel =
		selectedYears.length === 0
			? 'Año'
			: selectedYears.length === 1
				? `${selectedYears[0]}°`
				: selectedYears.map((year) => String(year)[0]).join(', ');

	function toggleYear(year: number) {
		onYearsChange(
			selectedYears.includes(year)
				? selectedYears.filter((selectedYear) => selectedYear !== year)
				: [...selectedYears, year].sort(),
		);
	}

	function handleWindowClick(event: MouseEvent) {
		if (yearMenu?.open && !yearMenu.contains(event.target as Node)) yearMenu.open = false;
	}

	function handleWindowKeydown(event: KeyboardEvent) {
		if (event.key !== 'Escape' || !yearMenu?.open) return;
		yearMenu.open = false;
		yearMenuButton?.focus();
	}
</script>

<svelte:window on:click={handleWindowClick} on:keydown={handleWindowKeydown} />

<aside class="neo-panel course-island flex h-full min-h-0 flex-col overflow-hidden p-3">
	<header class="flex items-center justify-between gap-2 px-1">
		<div class="mb-2">
			<h1 class="text-sm font-extrabold text-primary xl:text-[20px]">Cursos</h1>
			<p class="-mt-0.5 text-[11px] font-semibold text-muted">{termLabel}</p>
		</div>
		<div class="flex items-center gap-1">
			<div
				class={`inline-flex items-center gap-1 rounded-xl px-2.5 py-1.5 text-[12px] font-bold ${summary.conflictCount > 0 ? 'bg-warning-soft text-warning' : 'bg-success-soft text-success'}`}
				role="status"
				aria-live="polite"
			>
				{#if summary.conflictCount > 0}<svg
						class="h-4 w-4 stroke-current"
						viewBox="0 0 24 24"
						fill="none"
						stroke-width="2"
						stroke-linejoin="round"
						aria-hidden="true"
						><path d="M12 3 2.8 20h18.4L12 3Z" /><path d="M12 9v5M12 17h.01" /></svg
					>{/if}
				{summary.conflictCount > 0
					? `${summary.conflictCount} conflicto${summary.conflictCount === 1 ? '' : 's'}`
					: 'Sin conflictos'}
			</div>
			{#if showExports}<button
					class="neo-button grid h-8.5 w-9 place-items-center rounded-full text-secondary disabled:opacity-40"
					type="button"
					title="Previsualizar horario como imagen"
					aria-label="Previsualizar horario como imagen"
					disabled={summary.selectedCourses === 0}
					on:click={onPreviewImage}
				>
					<svg
						class="h-6 w-6 stroke-current"
						viewBox="0 0 24 24"
						fill="none"
						stroke-width="1.8"
						stroke-linecap="round"
						stroke-linejoin="round"
						aria-hidden="true"
						><rect x="3" y="5" width="18" height="14" rx="3" /><circle
							cx="9"
							cy="10"
							r="1.5"
						/><path d="m5.5 17 4.5-4 3 2.5 2.5-2 3 3.5" /></svg
					>
				</button>
				<button
					class="neo-button grid h-8.5 w-9 place-items-center text-secondary disabled:opacity-40"
					type="button"
					title="Descargar formato para Calendar (.ics)"
					aria-label="Descargar horario para Calendar en formato iCalendar"
					disabled={summary.selectedCourses === 0}
					on:click={onExportCalendar}
				>
					<svg
						class="h-6 w-6 stroke-current"
						viewBox="0 0 24 24"
						fill="none"
						stroke-width="1.8"
						stroke-linecap="round"
						stroke-linejoin="round"
						aria-hidden="true"
						><rect x="3" y="5" width="18" height="16" rx="3" /><path
							d="M8 3v4M16 3v4M3 10h18M8 14h3v3H8z"
						/></svg
					>
				</button>{/if}
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
		<details class="neo-control relative text-[16px]" bind:this={yearMenu}>
			<summary
				class="flex h-10 min-w-16 cursor-pointer list-none items-center justify-between gap-1 px-2 font-bold text-primary"
				aria-label="Filtrar por años"
				bind:this={yearMenuButton}
			>
				<span>{yearLabel}</span><svg
					class="h-4 w-4 shrink-0 stroke-current"
					viewBox="0 0 24 24"
					fill="none"
					stroke-width="2"
					aria-hidden="true"><path d="m7 9 5 5 5-5" /></svg
				>
			</summary>
			<div class="glass-panel absolute right-0 z-20 mt-1 w-36 rounded-xl p-2">
				{#each availableYears as year (year)}<label
						class="flex min-h-9 cursor-pointer items-center gap-2 rounded-lg px-2 text-xs font-bold text-primary hover:bg-surface-muted text-[14px]"
						><input
							class="h-4 w-4 accent-accent"
							type="checkbox"
							checked={selectedYears.includes(year)}
							on:change={() => toggleYear(year)}
						/>{getAcademicYearLabel(year)}</label
					>{/each}
				{#if selectedYears.length > 0}<button
						class="mt-1 min-h-8 w-full rounded-lg text-xs text-accent hover:bg-accent-soft font-bold"
						type="button"
						on:click={() => onYearsChange([])}>Todos</button
					>{/if}
			</div>
		</details>
		<button
			class="grid h-10 w-10 place-items-center rounded-xl bg-warning-soft text-warning transition hover:bg-warning/20 disabled:opacity-40"
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
	<div class="mt-0.75 flex items-center justify-between px-6 text-[12px] text-secondary">
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
					<CourseCard
						{bundle}
						{selectedGroups}
						{conflictingCourseIds}
						{onToggleGroup}
						{onOpenDetails}
					/>
				{/each}
			</div>
		{/if}
	</div>
</aside>
