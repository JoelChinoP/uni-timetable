<script lang="ts">
	import Modal from './Modal.svelte';
	import type { AcademicHour, Course, PlannerDay } from '../../types/planner';
	import type { ClassroomItem, GroupPayload, GroupSessionPayload } from '../../api/catalog';
	import { getCourseDisplayName, getDayLabel } from '../../utils/planner';

	export let course: Course;
	export let classrooms: ClassroomItem[] = [];
	export let academicHours: AcademicHour[] = [];
	export let busy = false;
	export let errorMessage = '';
	export let initialGroup: Course['groups'][number] | null = null;
	export let onSave: (payload: GroupPayload) => void;
	export let onClose: () => void;

	const days: PlannerDay[] = ['MONDAY', 'TUESDAY', 'WEDNESDAY', 'THURSDAY', 'FRIDAY', 'SATURDAY'];

	let name = initialGroup?.name ?? '';
	let classroomId: number | null = initialGroup?.classroomId ?? null;
	let rows: GroupSessionPayload[] = initialGroup?.sessions.map(
		({ day, startHourAcademic, durationHours }) => ({ day, startHourAcademic, durationHours }),
	) ?? [{ day: 'MONDAY', startHourAcademic: 1, durationHours: 2 }];

	$: availableClassrooms = classrooms.filter((classroom) => classroom.type === course.type);

	function addRow() {
		rows = [...rows, { day: 'MONDAY', startHourAcademic: 1, durationHours: 2 }];
	}

	function removeRow(index: number) {
		rows = rows.filter((_, row) => row !== index);
	}

	function submit() {
		onSave({
			courseId: course.id,
			name,
			classroomId,
			sessions: rows,
		});
	}

	const fieldClass = 'neo-control min-h-11 w-full px-3 py-2 text-sm text-primary outline-none';
	const labelClass = 'text-[10px] font-extrabold tracking-[0.16em] text-muted uppercase';
</script>

<Modal
	title={`${initialGroup ? 'Editar grupo' : 'Nuevo grupo'} · ${getCourseDisplayName(course)}`}
	{onClose}
>
	<form class="space-y-4" on:submit|preventDefault={submit}>
		<div class="grid gap-3 sm:grid-cols-2">
			<label class="flex flex-col gap-1.5">
				<span class={labelClass}>Grupo</span>
				<input
					class={fieldClass}
					bind:value={name}
					maxlength="5"
					placeholder="A"
					required
					type="text"
				/>
			</label>
			<label class="flex flex-col gap-1.5">
				<span class={labelClass}>Aula (opcional)</span>
				<select class={fieldClass} bind:value={classroomId}>
					<option value={null}>Sin aula</option>
					{#each availableClassrooms as classroom (classroom.id)}
						<option value={classroom.id}>{classroom.code}</option>
					{/each}
				</select>
			</label>
		</div>

		<fieldset>
			<legend class={labelClass}>Horarios del grupo</legend>
			<div class="mt-2 space-y-2">
				{#each rows as row, index (index)}
					<div class="neo-card grid gap-2 p-3 sm:grid-cols-[1fr_1fr_88px_44px] sm:items-end">
						<label class="flex flex-col gap-1">
							<span class={labelClass}>Día</span>
							<select class={fieldClass} bind:value={row.day}>
								{#each days as day (day)}
									<option value={day}>{getDayLabel(day)}</option>
								{/each}
							</select>
						</label>
						<label class="flex flex-col gap-1">
							<span class={labelClass}>Inicio</span>
							<select class={fieldClass} bind:value={row.startHourAcademic}>
								{#each academicHours as hour (hour.hourNumber)}
									<option value={hour.hourNumber}>
										{hour.hourNumber} · {hour.startTime}
									</option>
								{/each}
							</select>
						</label>
						<label class="flex flex-col gap-1">
							<span class={labelClass}>Bloques</span>
							<select class={fieldClass} bind:value={row.durationHours}>
								{#each [1, 2, 3, 4, 5, 6] as duration (duration)}
									<option value={duration}>{duration}</option>
								{/each}
							</select>
						</label>
						<button
							class="neo-button grid h-11 w-full place-items-center text-warning disabled:opacity-40 sm:w-11"
							type="button"
							aria-label="Quitar horario"
							disabled={rows.length === 1}
							on:click={() => removeRow(index)}
						>
							<svg
								class="h-4 w-4 stroke-current"
								viewBox="0 0 24 24"
								fill="none"
								stroke-width="2"
								stroke-linecap="round"
								aria-hidden="true"
							>
								<path d="M4 7h16M9 7V5a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v2m-9 0 1 13h8l1-13" />
							</svg>
						</button>
					</div>
				{/each}
			</div>
			<button
				class="mt-2 rounded-[10px] border border-dashed border-border-strong px-3 py-1.5 text-xs font-bold text-accent transition hover:border-accent/40 hover:bg-accent-soft"
				type="button"
				on:click={addRow}
			>
				+ Añadir horario
			</button>
		</fieldset>

		{#if errorMessage}
			<p
				class="rounded-2xl bg-warning-soft px-4 py-3 text-sm font-semibold text-warning"
				role="alert"
			>
				{errorMessage}
			</p>
		{/if}

		<button
			class="w-full rounded-[14px] bg-accent-strong px-4 py-3 text-sm font-bold text-white transition hover:bg-accent disabled:cursor-not-allowed disabled:opacity-60"
			disabled={busy}
			type="submit"
		>
			{busy ? 'Guardando…' : initialGroup ? 'Guardar cambios' : 'Guardar grupo con sus horarios'}
		</button>
	</form>
</Modal>
