export type PlannerDay = 'MONDAY' | 'TUESDAY' | 'WEDNESDAY' | 'THURSDAY' | 'FRIDAY' | 'SATURDAY';

export type SessionMode = 'THEORY' | 'LABORATORY';

export interface AcademicHour {
	hourNumber: number;
	startTime: string;
	endTime: string;
}

export interface TeacherSummary {
	id: number;
	fullName: string;
}

export interface GroupSession {
	id: number;
	title: string;
	day: PlannerDay;
	startHourAcademic: number;
	durationHours: number;
	mode: SessionMode;
	classroomLabel: string;
}

export interface CourseGroup {
	id: number;
	name: string;
	classroomId: number | null;
	classroomLabel: string;
	sessions: GroupSession[];
}

export interface Course {
	id: number;
	code: string;
	name: string;
	abbreviation: string;
	summary: string;
	credits: number | null;
	color: string;
	type: SessionMode;
	academicYear: number;
	theoryCourseId: number | null;
	teacher: TeacherSummary | null;
	groups: CourseGroup[];
}

export interface CourseBundle {
	key: string;
	theory: Course | null;
	laboratories: Course[];
}

export interface PlannerData {
	termLabel: string;
	days: PlannerDay[];
	academicHours: AcademicHour[];
	courses: Course[];
}

export interface SharedTimetable {
	id: string;
	selection: Record<string, number>;
	years: number[];
}

export interface SelectedCourseGroup {
	course: Course;
	group: CourseGroup;
}

export interface PlannerSummary {
	selectedCourses: number;
	conflictCount: number;
}

export interface PlannerEvent {
	id: string;
	courseId: number;
	groupId: number;
	groupName: string;
	sessionId: number;
	code: string;
	title: string;
	teacher: string;
	credits: number | null;
	day: PlannerDay;
	startHourAcademic: number;
	endHourAcademic: number;
	durationHours: number;
	mode: SessionMode;
	classroomId: number | null;
	classroomLabel: string;
	color: string;
	lane: number;
	laneCount: number;
	conflictIds: string[];
}

export interface PlannerConflict {
	id: string;
	day: PlannerDay;
	eventIds: [string, string];
}
