<script lang="ts">
	import type { Course } from '../types/planner';

	export let course: Course;
	export let selectedGroupId: number | null;
	export let onToggleGroup: (courseId: number, groupId: number) => void;
	export let onOpenDetails: (course: Course) => void;

	function getContrastTextColor(hexColor: string) {
		const normalized = hexColor.replace('#', '');
		if (normalized.length !== 6) {
			return '#0f172a';
		}

		const red = Number.parseInt(normalized.slice(0, 2), 16);
		const green = Number.parseInt(normalized.slice(2, 4), 16);
		const blue = Number.parseInt(normalized.slice(4, 6), 16);
		const luminance = (0.299 * red + 0.587 * green + 0.114 * blue) / 255;

		return luminance > 0.68 ? '#0f172a' : '#f8fafc';
	}

	$: cardStyle = selectedGroupId
		? `--course-accent:${course.color};border-color:color-mix(in srgb, var(--course-accent) 30%, var(--ui-border-strong));background:color-mix(in srgb, var(--course-accent) 5%, var(--ui-surface));`
		: `--course-accent:${course.color};`;
</script>

<article
	class="rounded-[14px] border border-border-subtle bg-surface px-2.5 py-2 transition duration-200 hover:border-border-strong"
	style={cardStyle}
>
	<div class="flex items-start justify-between gap-2">
		<button class="min-w-0 flex-1 text-left" type="button" on:click={() => onOpenDetails(course)}>
			<h3 class="truncate text-[13px] leading-4.5 font-semibold text-primary" title={course.name}>
				{course.name}
			</h3>
			<p class="mt-0.5 text-[10px] font-semibold tracking-wide text-muted uppercase">
				{course.abbreviation} · Año {course.academicYear}
				{course.type === 'LABORATORY' ? '· Lab' : ''}
			</p>
		</button>
		{#if selectedGroupId}
			<span
				class="mt-0.5 h-2 w-2 shrink-0 rounded-full"
				style={`background:${course.color};`}
				aria-hidden="true"
			></span>
		{/if}
	</div>

	<div class="mt-1.5 flex flex-wrap gap-1" aria-label={`${course.name} grupos`}>
		{#each course.groups as group (group.id)}
			<button
				class="min-w-[30px] rounded-md border border-border-subtle px-1.5 py-0.5 text-[11px] font-bold text-secondary transition duration-150 hover:border-accent/30 hover:text-primary"
				type="button"
				aria-pressed={group.id === selectedGroupId}
				style={group.id === selectedGroupId
					? `background:${course.color};border-color:transparent;color:${getContrastTextColor(course.color)};`
					: ''}
				on:click={() => onToggleGroup(course.id, group.id)}
			>
				{group.name}
			</button>
		{/each}
	</div>
</article>
