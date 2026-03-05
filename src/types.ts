import type { Application } from "$lib/server/db/schema";

export interface Stats {
  total: number;
  rejected: number;
  offers: number;
  inProgress: number;
}

export interface ApplicationWithStatus extends Application {
  latestStatus: string | null;
}

// This matches what the load function returns
export interface HomepageData {
  apps: ApplicationWithStatus[];
  stats: Stats;
}

export const eventTypes = [
		'APPLIED',
		'OA_SENT',
		'OA_DONE',
		'HIREVUE',
		'PHONE_SCREEN',
		'INTERVIEW',
		'OFFER',
		'REJECTED',
		'WITHDRAWN',
		'CLOSED'
	] as const;

export type EventType = (typeof eventTypes);