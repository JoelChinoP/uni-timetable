# Planner API Design

This document defines the backend endpoints needed to support the first planner view.
The current frontend uses a static mock file, but the payloads below are the target API shape.

## Read endpoints

### `GET /api/planner/dashboard`

Returns the full payload needed by the planner screen in a single request.

Query params:

- `term`
- `academicYear`
- `search`

Response:

- term metadata
- academic hours
- navigation metadata
- course catalog
- preselected groups for the active user

### `GET /api/academic-hours`

Returns the academic hour lookup table.

### `GET /api/teachers`

### `GET /api/teachers/:id`

### `POST /api/teachers`

### `PUT /api/teachers/:id`

### `DELETE /api/teachers/:id`

Teacher CRUD aligned with the `teachers` table.

### `GET /api/classrooms`

### `GET /api/classrooms/:id`

### `POST /api/classrooms`

### `PUT /api/classrooms/:id`

### `DELETE /api/classrooms/:id`

Classroom CRUD aligned with the `classrooms` table.

Filters:

- `type`
- `floor`
- `capacityMin`

### `GET /api/courses`

### `GET /api/courses/:id`

### `POST /api/courses`

### `PUT /api/courses/:id`

### `DELETE /api/courses/:id`

Course CRUD aligned with the `courses` table.

Filters:

- `search`
- `academicYear`
- `type`
- `teacherId`

### `GET /api/courses/:id/groups`

Returns all groups and schedule blocks for a course.

### `GET /api/groups`

### `GET /api/groups/:id`

### `POST /api/groups`

### `PUT /api/groups/:id`

### `DELETE /api/groups/:id`

Group CRUD aligned with the `groups` table.

Filters:

- `courseId`
- `classroomId`

### `GET /api/groups/:id/schedule`

### `POST /api/groups/:id/schedule`

### `PUT /api/groups/:id/schedule/:scheduleId`

### `DELETE /api/groups/:id/schedule/:scheduleId`

Schedule block CRUD aligned with the `schedule` table.

## Planner-specific actions

### `POST /api/planner/selections`

Persists the selected group per course for a user schedule draft.

Request body:

- `term`
- `selectedGroups`: map of course id to group id or `null`

### `POST /api/planner/conflicts/analyze`

Runs conflict detection on a draft selection and returns normalized conflict metadata for the UI.

Request body:

- `selectedGroups`

Response:

- conflict count
- conflicting sessions
- grouped conflict reasons by day

## Notes

- The frontend mock file mirrors the `GET /api/planner/dashboard` response.
- The dashboard payload is intentionally denormalized for UI speed.
- CRUD endpoints stay close to the relational schema, while planner endpoints are optimized for the screen.
