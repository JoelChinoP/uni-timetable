<script lang="ts">
	import type { Course, CourseBundle } from '../types/planner';

	export let bundle: CourseBundle;
	export let selectedGroups: Record<string, number> = {};
	export let onToggleGroup: (courseId: number, groupId: number) => void;
	export let onOpenDetails: (course: Course) => void;

	$: primary = bundle.theory ?? bundle.laboratories[0];
	$: related = [...(bundle.theory ? [bundle.theory] : []), ...bundle.laboratories];
	$: selectedCount = related.filter((course) => selectedGroups[String(course.id)]).length;

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
			<h3 class="truncate text-[13px] leading-5 font-extrabold text-primary" title={primary.name}>
				{primary.name}
			</h3>
			<p class="text-[10px] font-bold tracking-[0.12em] text-muted uppercase">
				{primary.abbreviation.replace(/-L$/, '')} · {primary.academicYear}° año
			</p>
		</button>
		<button
			class="grid h-10 w-10 shrink-0 place-items-center rounded-xl text-secondary transition hover:bg-surface-muted hover:text-accent"
			type="button"
			aria-label={`Ver detalles de ${primary.name}`}
			on:click={() => onOpenDetails(primary)}
		>
			<svg
				class="h-4 w-4 stroke-current"
				viewBox="0 0 24 24"
				fill="none"
				stroke-width="1.9"
				stroke-linecap="round"
				aria-hidden="true"><circle cx="12" cy="12" r="9" /><path d="M12 11v6M12 7.5h.01" /></svg
			>
		</button>
	</header>

	<div class="mt-2 grid gap-2">
		{#each related as course (course.id)}
			<section
				class="rounded-xl bg-surface-muted/70 p-2"
				aria-label={course.type === 'THEORY' ? 'Teoría' : 'Laboratorio'}
			>
				<div class="mb-1.5 flex items-center justify-between gap-2">
					<span class="text-[9px] font-extrabold tracking-[0.18em] text-secondary uppercase"
						>{course.type === 'THEORY'
							? 'Teoría'
							: bundle.laboratories.length > 1
								? `Laboratorio · ${course.abbreviation}`
								: 'Laboratorio'}</span
					>
					{#if selectedGroups[String(course.id)]}<span class="text-[9px] font-bold text-accent"
							>Seleccionado</span
						>{/if}
				</div>
				{#if course.groups.length === 0}
					<p class="text-[11px] text-muted">Sin grupos</p>
				{:else}
					<div class="flex flex-wrap gap-2">
						{#each course.groups as group (group.id)}
							<button
								class="neo-button min-h-10 min-w-10 px-2 text-xs font-extrabold text-secondary sm:min-h-9"
								type="button"
								aria-label={`${course.type === 'THEORY' ? 'Teoría' : 'Laboratorio'}, grupo ${group.name}`}
								aria-pressed={group.id === selectedGroups[String(course.id)]}
								style={group.id === selectedGroups[String(course.id)]
									? `background:${course.color};border-color:transparent;color:${getContrastTextColor(course.color)};`
									: ''}
								on:click={() => onToggleGroup(course.id, group.id)}>{group.name}</button
							>
						{/each}
					</div>
				{/if}
			</section>
		{/each}
	</div>
</article>
