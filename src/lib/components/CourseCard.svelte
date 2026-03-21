<script lang="ts">
	import type { AcademicHour, Course, CourseGroup } from '../types/planner';
	import { formatTimeRange, getGroupButtonTone, getModeLabel } from '../utils/planner';

	export let course: Course;
	export let academicHours: AcademicHour[] = [];
	export let selectedGroupId: number | null;
	export let onToggleGroup: (courseId: number, groupId: number) => void;
	export let onOpenDetails: (course: Course) => void;

	$: previewGroup =
		course.groups.find(({ id }) => id === selectedGroupId) ??
		course.groups.find(({ status }) => status === 'recommended') ??
		course.groups[0];

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

	function getCardStyle() {
		if (selectedGroupId) {
			return `--course-accent:${course.color};border-color:color-mix(in srgb, var(--course-accent) 26%, var(--ui-border-strong));background:color-mix(in srgb, var(--course-accent) 5%, var(--ui-surface));`;
		}

		return `--course-accent:${course.color};`;
	}

	function getGroupStyle(group: CourseGroup) {
		const tone = getGroupButtonTone(group, selectedGroupId);
		const contrastTextColor = getContrastTextColor(course.color);

		if (tone === 'selected') {
			return `--group-accent:${course.color};background:var(--group-accent);border-color:transparent;color:${contrastTextColor};`;
		}

		if (tone === 'recommended') {
			return `--group-accent:${course.color};background:color-mix(in srgb, var(--group-accent) 10%, var(--ui-surface));border-color:color-mix(in srgb, var(--group-accent) 26%, var(--ui-border));color:var(--group-accent);`;
		}

		return `--group-accent:${course.color};`;
	}
</script>

<article
	class="rounded-[18px] border border-border-subtle bg-surface p-3 shadow-card transition duration-200 hover:border-border-strong"
	style={getCardStyle()}
>
	<button class="block w-full text-left" type="button" on:click={() => onOpenDetails(course)}>
		<div class="space-y-2">
			{#each previewGroup?.sessions ?? [] as session (session.id)}
				<div class="rounded-[14px] bg-surface-muted px-3 py-2">
					<p class="text-[9px] font-extrabold uppercase tracking-[0.24em] text-secondary">
						{getModeLabel(session.mode)}
					</p>
					<h3 class="mt-1 text-sm font-semibold leading-5 text-primary">{session.title}</h3>
					<p class="mt-1 text-[11px] text-muted">
						{formatTimeRange(session.startHourAcademic, session.durationHours, academicHours)}
					</p>
				</div>
			{/each}
		</div>
	</button>

	<div class="mt-3 flex flex-wrap gap-1.5" aria-label={`${course.name} grupos`}>
		{#each course.groups as group (group.id)}
			<button
				class="min-w-[44px] rounded-lg border border-border-subtle px-2.5 py-1.5 text-xs font-bold text-secondary transition duration-200 hover:border-accent/30 hover:text-primary"
				type="button"
				aria-pressed={group.id === selectedGroupId}
				style={getGroupStyle(group)}
				on:click={() => onToggleGroup(course.id, group.id)}
			>
				{group.name}
			</button>
		{/each}
	</div>
</article>
