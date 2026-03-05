import { db } from '$lib/server/db';
import { applicationEvents, applications } from '$lib/server/db/schema';
import { eq } from 'drizzle-orm';
import { createEvents, type EventAttributes } from 'ics';

export const GET = async () => {
    const interviews = await db.select()
        .from(applicationEvents)
        .where(eq(applicationEvents.eventType, 'INTERVIEW'));

    const icalEvents: EventAttributes[] = interviews.map(event => ({
        start: new Date(event.eventDate).getTime(),
        duration: { hours: 1 },
        title: `Interview: ${event.details || 'Internship'}`,
        description: `Event details: ${event.details}`,
    }));

    const { error, value } = createEvents(icalEvents);

    return new Response(value, {
        headers: { 'Content-Type': 'text/calendar' }
    });
};