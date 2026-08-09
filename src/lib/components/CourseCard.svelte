<script lang="ts">
	import type { Course, CourseBundle } from '../types/planner';
	import { getCourseDisplayCode, getCourseDisplayName } from '../utils/planner';

	export let bundle: CourseBundle;
	export let selectedGroups: Record<string, number> = {};
	export let onToggleGroup: (courseId: number, groupId: number) => void;
	export let onOpenDetails: (course: Course) => void;
	export let conflictingCourseIds: Set<number> = new Set();

	$: primary = bundle.theory ?? bundle.laboratories[0];
	$: related = [...(bundle.theory ? [bundle.theory] : []), ...bundle.laboratories];
	$: selectedCount = related.filter((course) => selectedGroups[String(course.id)]).length;
	$: hasConflict = related.some((course) => conflictingCourseIds.has(course.id));

	function getContrastTextColor(hexColor: string) {
		const normalized = hexColor.replace('#', '');
		if (normalized.length !== 6) return '#0f172a';
		const [red, green, blue] = [0, 2, 4].map((start) =>
			Number.parseInt(normalized.slice(start, start + 2), 16),
		);
		return (0.299 * red + 0.587 * green + 0.114 * blue) / 255 > 0.62 ? '#0f172a' : '#ffffff';
	}
</script>

<article
	class={`relative overflow-hidden rounded-[16px] border bg-surface p-3 transition duration-200 ${
		selectedCount > 0 ? 'border-accent/40 shadow-card' : 'border-border-subtle'
	}`}
	style={`--course-accent:${primary.color};${selectedCount > 0 ? 'background:color-mix(in srgb,var(--course-accent) 6%,var(--ui-surface));' : ''}`}
>
	{#if selectedCount > 0}
		<span
			class="absolute inset-y-3 left-0 w-1 rounded-r-full"
			style={`background:${primary.color};`}
			aria-hidden="true"
		></span>
	{/if}
	<header class="flex items-start justify-between gap-2 pl-1">
		<button
			class="min-w-0 flex-1 rounded-lg text-left"
			type="button"
			on:click={() => onOpenDetails(primary)}
		>
			<h3
				class="truncate text-[14px] leading-5 font-extrabold text-primary"
				title={getCourseDisplayName(primary)}
			>
				{getCourseDisplayName(primary)}
			</h3>
			<p class="text-[10px] font-bold tracking-[0.12em] text-muted uppercase">
				{getCourseDisplayCode(primary)} · {primary.academicYear}° año
			</p>
		</button>
		<button
			class="grid h-8 w-8 shrink-0 place-items-center rounded-xl text-secondary transition hover:bg-surface-muted hover:text-accent"
			type="button"
			aria-label={`${hasConflict ? 'Ver conflicto y detalles' : 'Ver detalles'} de ${getCourseDisplayName(primary)}`}
			on:click={() => onOpenDetails(primary)}
		>
			{#if hasConflict}
				<svg
					class="h-6 w-6 stroke-warning"
					viewBox="0 0 24 24"
					fill="none"
					stroke-width="1.9"
					stroke-linecap="round"
					stroke-linejoin="round"
					aria-hidden="true"
					><path d="M12 3 2.8 20h18.4L12 3Z" /><path d="M12 9v5M12 17.2h.01" /></svg
				>
			{:else}
				<svg
					class="h-6 w-6 stroke-current"
					viewBox="0 0 24 24"
					fill="none"
					stroke-width="1.9"
					stroke-linecap="round"
					aria-hidden="true"><circle cx="12" cy="12" r="9" /><path d="M12 11v6M12 7.5h.01" /></svg
				>
			{/if}
		</button>
	</header>

	<div class="mt-1 flex flex-wrap gap-1">
		{#each related as course (course.id)}
			<section
				class={`relative min-w-0 rounded-md bg-surface-muted/70 p-1 ${related.length === 1 ? 'basis-full' : 'basis-[calc(50%_-_0.125rem)]'}`}
				aria-label={course.type === 'THEORY' ? 'Teoría' : 'Laboratorio'}
			>
				<div class="mb-1 flex items-center">
					<span
						class="min-w-0 text-[9px] font-extrabold tracking-[0.12em] text-secondary uppercase"
					>
						{course.type === 'THEORY' ? 'Teoría' : 'Laboratorio'}
					</span>

					{#if selectedGroups[String(course.id)]}
						<span
							class="absolute top-1 right-1 grid h-4 w-4 place-items-center rounded-full bg-accent text-accent-contrast"
							title="Seleccionado"
						>
							<svg
								class="h-3 w-3 stroke-current"
								viewBox="0 0 24 24"
								fill="none"
								stroke-width="3"
								stroke-linecap="round"
								stroke-linejoin="round"
								aria-hidden="true"
							>
								<path d="m6 12 4 4 8-9" />
							</svg>
							<span class="sr-only">Seleccionado</span>
						</span>
					{/if}
				</div>

				{#if course.groups.length === 0}
					<p class="text-[14px] text-muted">Sin grupos</p>
				{:else}
					<div class="flex flex-wrap gap-1">
						{#each course.groups as group (group.id)}
							<button
								class="neo-button min-h-8 min-w-8 text-xs font-extrabold text-secondary sm:min-h-8 sm:min-w-8"
								class:group-choice-selected={group.id === selectedGroups[String(course.id)]}
								type="button"
								aria-label={`${course.type === 'THEORY' ? 'Teoría' : 'Laboratorio'}, grupo ${group.name}`}
								aria-pressed={group.id === selectedGroups[String(course.id)]}
								style={group.id === selectedGroups[String(course.id)]
									? `background:${course.color};border-color:transparent;color:${getContrastTextColor(course.color)};font-weight:600;`
									: 'font-weight:600;'}
								on:click={() => onToggleGroup(course.id, group.id)}
							>
								{group.name}
							</button>
						{/each}
					</div>
				{/if}
			</section>
		{/each}
	</div>
</article>
