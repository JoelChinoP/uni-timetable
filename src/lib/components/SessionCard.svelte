<script lang="ts">
	import type { PlannerEvent } from '../types/planner';
	import { BOARD_EVENT_GAP, BOARD_HOUR_HEIGHT, getModeLabel } from '../utils/planner';

	export let event: PlannerEvent;
	export let timeLabel: string;
	export let focused = false;
	export let onOpen: (event: PlannerEvent) => void;

	$: top = (event.startHourAcademic - 1) * BOARD_HOUR_HEIGHT + BOARD_EVENT_GAP / 2;
	$: height = event.durationHours * BOARD_HOUR_HEIGHT - BOARD_EVENT_GAP;
	$: cardStyle = `--event-accent:${event.color};top:${top}px;left:calc((100% * ${event.lane} / ${event.laneCount}) + 6px);width:calc((100% / ${event.laneCount}) - 12px);height:${height}px;background:linear-gradient(180deg,color-mix(in srgb, var(--event-accent) 20%, var(--ui-surface)) 0%,color-mix(in srgb, var(--event-accent) 10%, var(--ui-surface)) 100%);border-color:color-mix(in srgb, var(--event-accent) 42%, var(--ui-border));box-shadow:0 16px 28px color-mix(in srgb, var(--event-accent) 16%, transparent);`;
</script>

<button
	class={`absolute flex min-h-[92px] flex-col overflow-hidden rounded-[22px] border border-l-[4px] p-3 text-left transition duration-200 hover:-translate-y-0.5 hover:shadow-panel ${
		focused ? 'shadow-panel' : ''
	} ${event.conflictIds.length > 0 ? 'ring-2 ring-warning/15' : ''}`}
	type="button"
	aria-label={`Abrir detalles de ${event.code}`}
	style={cardStyle}
	on:click={() => onOpen(event)}
>
	<div class="flex items-start gap-2">
		<span
			class="inline-flex min-h-9 min-w-9 items-center justify-center rounded-xl bg-surface px-2 text-sm font-extrabold shadow-sm"
			style={`color:${event.color};`}
		>
			{event.groupName}
		</span>

		<div class="min-w-0">
			<p class="text-[10px] font-extrabold uppercase tracking-[0.2em] text-secondary">
				{getModeLabel(event.mode)}
			</p>
			<h4 class="mt-1 line-clamp-2 text-sm font-bold leading-5 text-primary">
				{event.title}
			</h4>
		</div>
	</div>

	<p class="mt-2 line-clamp-1 text-xs text-secondary">{event.classroomLabel} - {event.teacher}</p>

	<div class="mt-auto flex items-end justify-between gap-2 pt-3">
		<span class="text-sm font-semibold text-secondary">{timeLabel}</span>

		{#if event.conflictIds.length > 0}
			<span class="rounded-full bg-warning/10 px-2.5 py-1 text-[10px] font-extrabold uppercase tracking-[0.18em] text-warning">
				Cruce
			</span>
		{/if}
	</div>
</button>
