import { db } from '$lib/server/db';
import { applications, applicationEvents, type Application, type ApplicationEvent } from '$lib/server/db/schema';
import { desc, eq } from 'drizzle-orm';
import type { ApplicationWithStatus, EventType } from '../types';
import { asc } from 'drizzle-orm';

export const load = async () => {
    const apps: Application[] = await db.select().from(applications).orderBy(desc(applications.id));
    const events: ApplicationEvent[] = await db.select().from(applicationEvents).orderBy(asc(applicationEvents.eventDate));

    const stats = {
        total: apps.length,
        rejected: events.filter(e => e.eventType === 'REJECTED').length,
        offers: events.filter(e => e.eventType === 'OFFER').length,
        inProgress: apps.length - events.filter(e => ['REJECTED', 'OFFER', 'WITHDRAWN'].includes(e.eventType)).length
    };

    const appsWithStatus: ApplicationWithStatus[] = apps.map((app)=> {
        const evt: string | null = events.findLast(e => e.applicationId === app.id)?.eventType ?? null;
        const appWithStatus: ApplicationWithStatus = { latestStatus: evt, ...app}
        return appWithStatus
    })

    return { apps: appsWithStatus, stats };
};