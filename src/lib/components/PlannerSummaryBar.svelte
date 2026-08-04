<script lang="ts">
	import type { PlannerSummary } from '../types/planner';
	import { formatSummaryHours } from '../utils/planner';

	export let summary: PlannerSummary;
</script>

<section class="grid gap-2 md:grid-cols-3" aria-label="Resumen del planner">
	<article
		class="rounded-[16px] border border-border-subtle bg-panel px-4 py-3 shadow-card backdrop-blur-xl"
	>
		<p class="text-[9px] font-extrabold tracking-[0.24em] text-muted uppercase">Cursos</p>
		<div class="mt-1.5 flex items-end justify-between gap-3">
			<strong class="text-2xl leading-none font-bold text-primary">{summary.selectedCourses}</strong
			>
			<span class="text-xs text-secondary">seleccionados</span>
		</div>
	</article>

	<article
		class="rounded-[16px] border border-border-subtle bg-panel px-4 py-3 shadow-card backdrop-blur-xl"
	>
		<p class="text-[9px] font-extrabold tracking-[0.24em] text-muted uppercase">Horas</p>
		<div class="mt-1.5 flex items-end justify-between gap-3">
			<strong class="text-2xl leading-none font-bold text-primary">
				{formatSummaryHours(summary.weeklyHours)}
			</strong>
			<span class="text-xs text-secondary">por semana</span>
		</div>
	</article>

	<article
		class={`rounded-[16px] border px-4 py-3 shadow-card backdrop-blur-xl ${
			summary.conflictCount > 0
				? 'border-warning/20 bg-warning-soft'
				: 'border-border-subtle bg-panel'
		}`}
	>
		<p class="text-[9px] font-extrabold tracking-[0.24em] text-muted uppercase">Conflictos</p>
		<div class="mt-1.5 flex items-end justify-between gap-3">
			<strong
				class={`text-2xl leading-none font-bold ${summary.conflictCount > 0 ? 'text-warning' : 'text-primary'}`}
			>
				{summary.conflictCount}
			</strong>
			<span class="text-xs text-secondary"
				>{summary.conflictCount > 0 ? 'activos' : 'sin cruces'}</span
			>
		</div>
	</article>
</section>
