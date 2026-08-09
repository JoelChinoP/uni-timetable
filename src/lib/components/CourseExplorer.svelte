<script lang="ts">
	import CourseCard from './CourseCard.svelte';
	import { groupRelatedCourses } from '../utils/planner';
	import type { Course } from '../types/planner';
	import type { ClassroomItem } from '../api/catalog';

	export let termLabel: string;
	export let searchQuery = '';
	export let selectedYear: number | null = null;
	export let availableYears: number[] = [];
	export let classrooms: ClassroomItem[] = [];
	export let selectedClassroomId: number | null = null;
	export let courses: Course[] = [];
	export let selectedGroups: Record<string, number> = {};
	export let summary: { selectedCourses: number; conflictCount: number };
	export let onSearchChange: (value: string) => void;
	export let onYearChange: (year: number | null) => void;
	export let onClassroomChange: (classroomId: number | null) => void;
	export let onToggleGroup: (courseId: number, groupId: number) => void;
	export let onClearSelection: () => void;
	export let onOpenDetails: (course: Course) => void;
	export let conflictingCourseIds: Set<number> = new Set();
	export let onPreviewImage: () => void = () => {};
	export let onExportCalendar: () => void = () => {};
	export let calendarEnabled = false;
	export let exportingCalendar = false;

	$: bundles = groupRelatedCourses(courses);
</script>

<aside class="neo-panel course-island flex h-full min-h-0 flex-col overflow-hidden p-3">
	<header class="flex items-center justify-between gap-2 px-1">
		<div class="mb-2">
			<h1 class="text-sm font-extrabold text-primary xl:text-[18px]">Cursos</h1>
			<p class="-mt-0.5 text-[11px] font-semibold text-muted">{termLabel}</p>
		</div>
		<div class="flex items-center gap-1">
			<div
				class={`inline-flex items-center gap-1 rounded-xl px-2.5 py-1.5 text-[10px] font-bold ${summary.conflictCount > 0 ? 'bg-warning-soft text-warning' : 'bg-success-soft text-success'}`}
				role="status"
				aria-live="polite"
			>
				{#if summary.conflictCount > 0}<svg
						class="h-3.5 w-3.5 stroke-current"
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
			<button
				class="neo-button grid h-10 w-10 place-items-center text-secondary disabled:opacity-40"
				type="button"
				title="Previsualizar horario como imagen"
				aria-label="Previsualizar horario como imagen"
				disabled={summary.selectedCourses === 0}
				on:click={onPreviewImage}
			>
				<svg
					class="h-5 w-5 stroke-current"
					viewBox="0 0 24 24"
					fill="none"
					stroke-width="1.8"
					stroke-linecap="round"
					stroke-linejoin="round"
					aria-hidden="true"
					><rect x="3" y="5" width="18" height="14" rx="3" /><circle cx="9" cy="10" r="1.5" /><path
						d="m5.5 17 4.5-4 3 2.5 2.5-2 3 3.5"
					/></svg
				>
			</button>
			<button
				class="neo-button grid h-10 w-10 place-items-center text-secondary disabled:opacity-40"
				type="button"
				title={calendarEnabled
					? 'Exportar a Google Calendar'
					: 'Inicia sesión para exportar a Calendar'}
				aria-label={calendarEnabled
					? 'Exportar a Google Calendar'
					: 'Inicia sesión para exportar a Calendar'}
				disabled={!calendarEnabled || summary.selectedCourses === 0 || exportingCalendar}
				on:click={onExportCalendar}
			>
				<svg
					class="h-5 w-5 stroke-current"
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
			</button>
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
	<label class="neo-control mt-1 grid px-2 text-[16px]" aria-label="Filtrar por aula">
		<select
			class="h-9 min-w-0 bg-transparent px-1 font-bold text-primary outline-none"
			value={selectedClassroomId ?? ''}
			on:change={(event) =>
				onClassroomChange(
					(event.currentTarget as HTMLSelectElement).value
						? Number((event.currentTarget as HTMLSelectElement).value)
						: null,
				)}
		>
			<option value="">Todas las aulas</option>
			{#each classrooms as classroom (classroom.id)}<option value={classroom.id}
					>{classroom.code} · {classroom.type === 'THEORY' ? 'Teoría' : 'Laboratorio'}</option
				>{/each}
		</select>
	</label>

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
