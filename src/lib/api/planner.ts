import type { PlannerDashboard } from '../types/planner';

const MOCK_PLANNER_ENDPOINT = '/mocks/planner-dashboard.json';

export async function loadPlannerDashboard(): Promise<PlannerDashboard> {
	const response = await fetch(MOCK_PLANNER_ENDPOINT);
	if (!response.ok) {
		throw new Error('Unable to load planner dashboard mock data.');
	}

	const payload = (await response.json()) as PlannerDashboard;
	return payload;

	/*
  const response = await fetch('/api/planner/dashboard');
  if (!response.ok) {
    throw new Error('Unable to load planner dashboard data.');
  }

  return (await response.json()) as PlannerDashboard;
  */
}
