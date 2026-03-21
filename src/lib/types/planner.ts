export type PlannerDay = 'MONDAY' | 'TUESDAY' | 'WEDNESDAY' | 'THURSDAY' | 'FRIDAY' | 'SATURDAY';

export type SessionMode = 'THEORY' | 'LABORATORY';

export interface AcademicHour {
	hourNumber: number;
	startTime: string;
	endTime: string;
}

export interface NavigationItem {
	id: string;
	label: string;
	active: boolean;
}

export interface PlannerTab {
	id: string;
	label: string;
	active: boolean;
}

export interface PlannerUser {
	name: string;
	initials: string;
	avatarLabel: string;
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
	status: 'recommended' | 'available';
	sessions: GroupSession[];
}

export interface Course {
	id: number;
	code: string;
	name: string;
	abbreviation: string;
	summary: string;
	credits: number;
	academicYear: number;
	color: string;
	teacher: TeacherSummary;
	groups: CourseGroup[];
}

export interface PlannerDashboard {
	termLabel: string;
	boardTitle: string;
	boardSubtitle: string;
	days: PlannerDay[];
	academicHours: AcademicHour[];
	navigation: NavigationItem[];
	tabs: PlannerTab[];
	user: PlannerUser;
	courses: Course[];
	selectedGroups: Record<string, number | null>;
}

export interface SelectedCourseGroup {
	course: Course;
	group: CourseGroup;
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
	credits: number;
	day: PlannerDay;
	startHourAcademic: number;
	endHourAcademic: number;
	durationHours: number;
	mode: SessionMode;
	classroomLabel: string;
	color: string;
	lane: number;
	laneCount: number;
	conflictIds: string[];
}

export interface PlannerSummary {
	selectedCourses: number;
	weeklyHours: number;
	conflictCount: number;
}

export interface PlannerConflict {
	id: string;
	day: PlannerDay;
	eventIds: string[];
}
