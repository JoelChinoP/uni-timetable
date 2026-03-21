<script lang="ts">
	import type { Course } from '../types/planner';
	import CourseCard from './CourseCard.svelte';

	export let termLabel: string;
	export let searchQuery = '';
	export let courses: Course[] = [];
	export let selectedGroups: Record<string, number | null> = {};
	export let onSearchChange: (value: string) => void;
	export let onToggleGroup: (courseId: number, groupId: number) => void;
	export let onOpenDetails: (course: Course) => void;
</script>

<aside
	class="flex min-h-[760px] flex-col rounded-[32px] border border-border-subtle bg-panel p-5 shadow-panel backdrop-blur-xl xl:min-h-0"
>
	<div class="space-y-4 border-b border-border-subtle pb-5">
		<div class="flex items-start justify-between gap-4">
			<div class="min-w-0">
				<p class="text-[11px] font-extrabold uppercase tracking-[0.24em] text-accent">
					Explorador
				</p>
				<h2 class="mt-2 font-display text-3xl leading-none text-primary">Turno por materia</h2>
				<p class="mt-2 text-sm text-secondary">{termLabel}</p>
			</div>

			<div class="rounded-[20px] border border-border-subtle bg-surface px-4 py-3 text-right shadow-card">
				<p class="text-[10px] font-extrabold uppercase tracking-[0.22em] text-muted">Visibles</p>
				<p class="mt-1 text-2xl font-bold text-primary">{courses.length}</p>
			</div>
		</div>

		<label
			class="flex items-center gap-3 rounded-[22px] border border-border-subtle bg-surface px-4 py-3 shadow-card transition duration-200 focus-within:border-accent/35 focus-within:ring-4 focus-within:ring-accent-soft"
			aria-label="Buscar cursos"
		>
			<svg
				class="h-5 w-5 shrink-0 stroke-muted"
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
				placeholder="Busca por nombre, codigo o tema..."
				value={searchQuery}
				on:input={(event) => onSearchChange((event.currentTarget as HTMLInputElement).value)}
			/>
		</label>

		<div class="flex flex-wrap gap-2">
			<span
				class="rounded-full bg-accent-soft px-3 py-1.5 text-[10px] font-extrabold uppercase tracking-[0.22em] text-accent"
			>
				{courses.length} cursos visibles
			</span>
			<span
				class="rounded-full bg-surface px-3 py-1.5 text-[10px] font-extrabold uppercase tracking-[0.22em] text-muted shadow-card"
			>
				1 grupo por curso
			</span>
		</div>
	</div>

	<div class="mt-5 min-h-0 flex-1 overflow-y-auto pr-1">
		<div class="space-y-3">
			{#if courses.length === 0}
				<div class="rounded-[28px] border border-dashed border-border-strong bg-surface p-5 shadow-card">
					<h3 class="text-lg font-bold text-primary">No encontramos coincidencias</h3>
					<p class="mt-2 text-sm leading-6 text-secondary">
						Prueba con otra palabra clave o limpia la busqueda para recuperar la lista.
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

	<div class="mt-5 rounded-[24px] border border-border-subtle bg-surface px-4 py-4 shadow-card">
		<p class="text-[10px] font-extrabold uppercase tracking-[0.22em] text-muted">Consejo</p>
		<p class="mt-2 text-sm leading-6 text-secondary">
			Selecciona una sola seccion por curso para que el horario se mantenga limpio y los cruces
			sean mas faciles de revisar.
		</p>
	</div>
</aside>
