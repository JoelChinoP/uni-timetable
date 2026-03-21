<script lang="ts">
	import type { PlannerSummary } from '../types/planner';
	import { formatSummaryHours } from '../utils/planner';

	export let summary: PlannerSummary;
</script>

<section class="grid gap-4 md:grid-cols-3" aria-label="Resumen del planner">
	<article
		class="rounded-[28px] border border-border-subtle bg-panel px-5 py-4 shadow-card backdrop-blur-xl"
	>
		<p class="text-[10px] font-extrabold uppercase tracking-[0.24em] text-muted">
			Cursos seleccionados
		</p>
		<div class="mt-3 flex items-end justify-between gap-3">
			<strong class="font-display text-4xl leading-none text-primary">
				{summary.selectedCourses}
			</strong>
			<span class="rounded-full bg-accent-soft px-3 py-1 text-xs font-bold text-accent">
				Activos
			</span>
		</div>
	</article>

	<article
		class="rounded-[28px] border border-border-subtle bg-panel px-5 py-4 shadow-card backdrop-blur-xl"
	>
		<p class="text-[10px] font-extrabold uppercase tracking-[0.24em] text-muted">
			Carga semanal
		</p>
		<div class="mt-3 flex items-end justify-between gap-3">
			<strong class="font-display text-4xl leading-none text-primary">
				{formatSummaryHours(summary.weeklyHours)}
			</strong>
			<span class="rounded-full bg-surface px-3 py-1 text-xs font-bold text-secondary">
				Horas
			</span>
		</div>
	</article>

	<article
		class={`rounded-[28px] border px-5 py-4 shadow-card backdrop-blur-xl ${
			summary.conflictCount > 0
				? 'border-warning/20 bg-warning-soft'
				: 'border-border-subtle bg-panel'
		}`}
	>
		<p class="text-[10px] font-extrabold uppercase tracking-[0.24em] text-muted">Cruces</p>
		<div class="mt-3 flex items-end justify-between gap-3">
			<strong
				class={`font-display text-4xl leading-none ${
					summary.conflictCount > 0 ? 'text-warning' : 'text-primary'
				}`}
			>
				{summary.conflictCount > 0 ? summary.conflictCount : '0'}
			</strong>
			<span
				class={`rounded-full px-3 py-1 text-xs font-bold ${
					summary.conflictCount > 0
						? 'bg-warning/10 text-warning'
						: 'bg-surface text-secondary'
				}`}
			>
				{summary.conflictCount > 0 ? 'Revisar' : 'Todo claro'}
			</span>
		</div>
	</article>
</section>
