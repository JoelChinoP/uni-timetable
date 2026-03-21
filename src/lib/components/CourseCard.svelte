<script lang="ts">
	import type { Course, CourseGroup } from '../types/planner';
	import { getGroupButtonTone } from '../utils/planner';

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

	function getCardStyle() {
		if (selectedGroupId) {
			return `--course-accent:${course.color};border-color:color-mix(in srgb, var(--course-accent) 28%, var(--ui-border-strong));background:color-mix(in srgb, var(--course-accent) 5%, var(--ui-surface));`;
		}

		return `--course-accent:${course.color};`;
	}

	function getGroupStyle(group: CourseGroup) {
		const tone = getGroupButtonTone(group, selectedGroupId);
		const contrastTextColor = getContrastTextColor(course.color);

		if (tone === 'selected') {
			return `--group-accent:${course.color};background:var(--group-accent);border-color:transparent;color:${contrastTextColor};box-shadow:0 12px 22px color-mix(in srgb, var(--group-accent) 28%, transparent);`;
		}

		if (tone === 'recommended') {
			return `--group-accent:${course.color};background:color-mix(in srgb, var(--group-accent) 12%, var(--ui-surface));border-color:color-mix(in srgb, var(--group-accent) 28%, var(--ui-border));color:var(--group-accent);`;
		}

		return `--group-accent:${course.color};`;
	}
</script>

<article
	class="rounded-[28px] border border-border-subtle bg-surface p-4 shadow-card transition duration-200 hover:-translate-y-0.5 hover:shadow-panel"
	style={getCardStyle()}
>
	<div class="flex items-start gap-3">
		<span class="mt-1 h-3 w-3 shrink-0 rounded-full" style={`background:${course.color};`}></span>

		<div class="min-w-0 flex-1">
			<button class="block w-full text-left" type="button" on:click={() => onOpenDetails(course)}>
				<div class="flex items-start justify-between gap-3">
					<div class="min-w-0">
						<p
							class="text-[10px] font-extrabold uppercase tracking-[0.22em]"
							style={`color:${course.color};`}
						>
							{course.code}
						</p>
						<h3 class="mt-1 line-clamp-2 text-base font-bold leading-6 text-primary">
							{course.name}
						</h3>
					</div>

					<div class="flex shrink-0 flex-col items-end gap-2">
						<span
							class="rounded-full bg-surface-muted px-3 py-1 text-[10px] font-extrabold uppercase tracking-[0.2em] text-muted"
						>
							{course.credits} cr
						</span>
						<span
							class={`rounded-full px-3 py-1 text-[10px] font-extrabold uppercase tracking-[0.2em] ${
								selectedGroupId ? 'bg-accent-soft text-accent' : 'bg-surface-muted text-muted'
							}`}
						>
							{selectedGroupId ? 'Activo' : 'Pendiente'}
						</span>
					</div>
				</div>

				<p class="mt-3 line-clamp-2 text-sm leading-6 text-secondary">{course.summary}</p>
			</button>

			<div class="mt-3 flex flex-wrap gap-2">
				<span class="rounded-full bg-surface-muted px-3 py-1.5 text-xs text-secondary">
					{course.teacher.fullName}
				</span>
				<span class="rounded-full bg-surface-muted px-3 py-1.5 text-xs text-secondary">
					Ano {course.academicYear}
				</span>
				<span class="rounded-full bg-surface-muted px-3 py-1.5 text-xs text-secondary">
					{course.groups.length} grupos
				</span>
			</div>

			<div class="mt-4 flex flex-wrap gap-2" aria-label={`${course.code} grupos`}>
				{#each course.groups as group (group.id)}
					<button
						class="min-w-[56px] rounded-xl border border-border-subtle px-3 py-2 text-sm font-bold text-secondary transition duration-200 hover:-translate-y-0.5 hover:border-accent/30 hover:text-primary"
						type="button"
						aria-pressed={group.id === selectedGroupId}
						style={getGroupStyle(group)}
						on:click={() => onToggleGroup(course.id, group.id)}
					>
						{group.name}
					</button>
				{/each}
			</div>
		</div>
	</div>
</article>
