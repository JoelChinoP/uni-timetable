<script lang="ts">
	import SessionCard from './SessionCard.svelte';
	import type {
		AcademicHour,
		PlannerConflict,
		PlannerDay,
		PlannerEvent,
		PlannerTab,
	} from '../types/planner';
	import { BOARD_HOUR_HEIGHT, formatTimeRange, getDayLabel } from '../utils/planner';

	export let boardTitle: string;
	export let boardSubtitle: string;
	export let tabs: PlannerTab[] = [];
	export let days: PlannerDay[] = [];
	export let academicHours: AcademicHour[] = [];
	export let events: PlannerEvent[] = [];
	export let conflicts: PlannerConflict[] = [];
	export let onOpenEvent: (event: PlannerEvent) => void;

	$: featuredDay = days.reduce((activeDay, day) => {
		const currentTotal = events.filter((event) => event.day === day).length;
		const activeTotal = events.filter((event) => event.day === activeDay).length;
		return currentTotal > activeTotal ? day : activeDay;
	}, days[0]);

	$: eventsByDay = days.reduce<Record<PlannerDay, PlannerEvent[]>>(
		(grouped, day) => {
			grouped[day] = events.filter((event) => event.day === day);
			return grouped;
		},
		{} as Record<PlannerDay, PlannerEvent[]>,
	);

	$: dayColumnsStyle = `grid-template-columns: repeat(${days.length}, minmax(180px, 1fr));`;
	$: boardMinWidth = 88 + days.length * 190;
	$: boardRowsHeight = academicHours.length * BOARD_HOUR_HEIGHT;
	$: rowsStyle = `grid-auto-rows:${BOARD_HOUR_HEIGHT}px;min-height:${boardRowsHeight}px;`;
</script>

<section
	class="flex min-h-[760px] flex-col rounded-[32px] border border-border-subtle bg-panel p-4 shadow-panel backdrop-blur-xl lg:p-5 xl:min-h-0"
>
	<div class="border-b border-border-subtle pb-4">
		<div class="flex flex-col gap-4 xl:flex-row xl:items-start xl:justify-between">
			<div class="space-y-2">
				<div class="flex flex-wrap gap-2">
					<span
						class="rounded-full bg-accent-soft px-3 py-1.5 text-[10px] font-extrabold uppercase tracking-[0.22em] text-accent"
					>
						Horario semanal
					</span>
					<span
						class={`rounded-full px-3 py-1.5 text-[10px] font-extrabold uppercase tracking-[0.22em] ${
							conflicts.length > 0
								? 'bg-warning-soft text-warning'
								: 'bg-surface text-secondary'
						}`}
					>
						{conflicts.length > 0 ? `${conflicts.length} cruces` : 'Sin cruces'}
					</span>
				</div>

				<div>
					<h2 class="font-display text-3xl leading-none text-primary sm:text-[2.6rem]">
						{boardTitle}
					</h2>
					<p class="mt-2 max-w-3xl text-sm leading-6 text-secondary">{boardSubtitle}</p>
				</div>
			</div>

			<div class="flex flex-col gap-3 xl:items-end">
				{#if tabs.length > 0}
					<div
						class="flex flex-wrap items-center gap-2 rounded-full border border-border-subtle bg-surface p-1.5 shadow-card"
						aria-label="Vista del tablero"
					>
						{#each tabs as tab (tab.id)}
							<button
								class={`rounded-full px-4 py-2 text-sm font-bold transition duration-200 ${
									tab.active
										? 'bg-accent-soft text-accent'
										: 'text-secondary hover:bg-surface-muted hover:text-primary'
								}`}
								type="button"
							>
								{tab.label}
							</button>
						{/each}
					</div>
				{/if}

				<div class="flex flex-wrap gap-2">
					<span
						class="rounded-full bg-surface px-3 py-1.5 text-[10px] font-extrabold uppercase tracking-[0.22em] text-muted shadow-card"
					>
						{days.length} dias
					</span>
					<span
						class="rounded-full bg-surface px-3 py-1.5 text-[10px] font-extrabold uppercase tracking-[0.22em] text-muted shadow-card"
					>
						{events.length} bloques
					</span>
				</div>
			</div>
		</div>

		{#if conflicts.length > 0}
			<div
				class="mt-4 flex flex-col gap-1 rounded-[22px] border border-warning/20 bg-warning-soft px-4 py-3 sm:flex-row sm:items-center sm:justify-between"
			>
				<p class="text-sm font-bold text-warning">Hay cruces entre secciones seleccionadas.</p>
				<p class="text-sm text-secondary">
					Los bloques en conflicto quedan resaltados directamente en el horario.
				</p>
			</div>
		{/if}
	</div>

	<div class="mt-4 min-h-0 overflow-hidden rounded-[30px] border border-border-subtle bg-surface shadow-card">
		<div class="min-h-0 overflow-auto">
			<div
				class="grid"
				style={`grid-template-columns:88px minmax(0,1fr);min-width:${boardMinWidth}px;`}
			>
				<div
					class="flex items-center justify-center border-b border-r border-border-subtle bg-surface-muted px-3 py-4 text-[10px] font-extrabold uppercase tracking-[0.24em] text-muted"
				>
					Hora
				</div>

				<div class="grid border-b border-border-subtle bg-surface-muted" style={dayColumnsStyle}>
					{#each days as day (day)}
						<div
							class={`flex items-center justify-center border-r border-border-subtle px-4 py-4 text-base font-bold last:border-r-0 ${
								day === featuredDay
									? 'border-b-2 border-b-accent bg-accent-soft text-accent'
									: 'text-primary'
							}`}
						>
							{getDayLabel(day)}
						</div>
					{/each}
				</div>

				<div class="grid border-r border-border-subtle bg-surface-muted" style={rowsStyle}>
					{#each academicHours as hour (hour.hourNumber)}
						<div
							class="border-b border-border-subtle px-3 py-2 text-right text-sm font-medium text-secondary last:border-b-0"
						>
							{hour.startTime}
						</div>
					{/each}
				</div>

				<div class="grid" style={dayColumnsStyle}>
					{#each days as day (day)}
						<div
							class="relative grid border-r border-border-subtle last:border-r-0"
							style={rowsStyle}
						>
							{#each academicHours as hour (hour.hourNumber)}
								<div
									class={`border-b border-border-subtle last:border-b-0 ${
										day === featuredDay ? 'bg-accent-soft' : ''
									}`}
								></div>
							{/each}

							<div class="absolute inset-0">
								{#each eventsByDay[day] ?? [] as event (event.id)}
									<SessionCard
										{event}
										focused={event.conflictIds.length > 0}
										timeLabel={formatTimeRange(
											event.startHourAcademic,
											event.durationHours,
											academicHours,
										)}
										onOpen={onOpenEvent}
									/>
								{/each}
							</div>
						</div>
					{/each}
				</div>
			</div>
		</div>
	</div>
</section>
