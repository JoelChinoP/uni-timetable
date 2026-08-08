<script lang="ts">
	import Modal from './Modal.svelte';
	import type { Course, SessionMode } from '../../types/planner';
	import type { CoursePayload, TeacherItem } from '../../api/catalog';

	export let courses: Course[] = [];
	export let teachers: TeacherItem[] = [];
	export let busy = false;
	export let errorMessage = '';
	export let initialCourse: Course | null = null;
	export let onSave: (payload: CoursePayload) => void;
	export let onClose: () => void;

	let name = initialCourse?.name ?? '';
	let abbreviation = initialCourse?.abbreviation.replace(/-L$/, '') ?? '';
	let type: SessionMode = initialCourse?.type ?? 'THEORY';
	let academicYear = initialCourse?.academicYear ?? 1;
	let theoryCourseId: number | null = initialCourse?.theoryCourseId ?? null;
	let teacherId: number | null = initialCourse?.teacher?.id ?? null;
	let credits: number | null = initialCourse?.credits ?? null;
	let color = initialCourse?.color ?? '#2563eb';

	$: theoryCourses = courses.filter((course) => course.type === 'THEORY');

	function submit() {
		onSave({
			name,
			abbreviation,
			type,
			academicYear,
			theoryCourseId: type === 'LABORATORY' ? theoryCourseId : null,
			teacherId,
			credits: credits || null,
			color,
		});
	}

	const fieldClass = 'neo-control min-h-11 w-full px-3 py-2 text-sm text-primary outline-none';
	const labelClass = 'text-[10px] font-extrabold tracking-[0.16em] text-muted uppercase';
</script>

<Modal title={initialCourse ? `Editar · ${initialCourse.name}` : 'Nuevo curso'} {onClose}>
	<form class="grid gap-3 sm:grid-cols-2" on:submit|preventDefault={submit}>
		<label class="flex flex-col gap-1.5 sm:col-span-2">
			<span class={labelClass}>Nombre</span>
			<input class={fieldClass} bind:value={name} required type="text" />
		</label>
		<label class="flex flex-col gap-1.5">
			<span class={labelClass}>Sigla</span>
			<input
				class={fieldClass}
				bind:value={abbreviation}
				maxlength="10"
				placeholder="DSOO"
				required
				type="text"
			/>
		</label>
		<label class="flex flex-col gap-1.5">
			<span class={labelClass}>Modalidad</span>
			<select class={fieldClass} bind:value={type} disabled={!!initialCourse}>
				<option value="THEORY">Teoría</option>
				<option value="LABORATORY">Laboratorio</option>
			</select>
		</label>
		<label class="flex flex-col gap-1.5">
			<span class={labelClass}>Año</span>
			<select class={fieldClass} bind:value={academicYear}>
				{#each [1, 2, 3, 4, 5] as year (year)}
					<option value={year}>{year}.º año</option>
				{/each}
			</select>
		</label>
		<label class="flex flex-col gap-1.5">
			<span class={labelClass}>Color</span>
			<input class="neo-control h-11 w-full cursor-pointer p-1.5" bind:value={color} type="color" />
		</label>
		<label class="flex flex-col gap-1.5">
			<span class={labelClass}>Créditos (opcional)</span>
			<input class={fieldClass} bind:value={credits} min="1" max="40" type="number" />
		</label>
		{#if type === 'LABORATORY'}
			<label class="flex flex-col gap-1.5 sm:col-span-2">
				<span class={labelClass}>Curso de teoría al que pertenece</span>
				<select class={fieldClass} bind:value={theoryCourseId} required>
					<option value={null} disabled>Selecciona…</option>
					{#each theoryCourses as course (course.id)}
						<option value={course.id}>{course.name}</option>
					{/each}
				</select>
			</label>
		{/if}
		<label class="flex flex-col gap-1.5 sm:col-span-2">
			<span class={labelClass}>Docente (opcional)</span>
			<select class={fieldClass} bind:value={teacherId}>
				<option value={null}>Sin docente</option>
				{#each teachers as teacher (teacher.id)}
					<option value={teacher.id}>{teacher.fullName}</option>
				{/each}
			</select>
		</label>

		{#if errorMessage}
			<p
				class="rounded-2xl bg-warning-soft px-4 py-3 text-sm font-semibold text-warning sm:col-span-2"
				role="alert"
			>
				{errorMessage}
			</p>
		{/if}

		<button
			class="rounded-[14px] bg-accent-strong px-4 py-3 text-sm font-bold text-white transition hover:bg-accent disabled:cursor-not-allowed disabled:opacity-60 sm:col-span-2"
			disabled={busy}
			type="submit"
		>
			{busy ? 'Guardando…' : initialCourse ? 'Guardar cambios' : 'Guardar curso'}
		</button>
	</form>
</Modal>
