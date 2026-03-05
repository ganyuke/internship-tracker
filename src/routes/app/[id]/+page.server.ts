import { db } from '$lib/server/db';
import { applications, applicationEvents } from '$lib/server/db/schema';
import { error, redirect } from '@sveltejs/kit';
import { eq, asc } from 'drizzle-orm';
import type { PageServerLoad, Actions } from './$types';
import { eventTypes, type EventType } from '../../../types';

function isEventType(v: unknown): v is EventType {
    return typeof v === 'string' && (eventTypes as readonly string[]).includes(v);
}

export const load: PageServerLoad = async ({ params }) => {
    const appId = parseInt(params.id);

    const app = await db.query.applications.findFirst({
        where: eq(applications.id, appId)
    });

    if (!app) throw error(404, 'Application not found');

    const events = await db.query.applicationEvents.findMany({
        where: eq(applicationEvents.applicationId, appId),
        orderBy: [asc(applicationEvents.eventDate)]
    });

    return { app, events };
};

export const actions: Actions = {
    // Action to update the main application details
    updateDetails: async ({ request, params }) => {
        const appId = parseInt(params.id);
        const data = await request.formData();

        await db.update(applications)
            .set({
                company: data.get('company') as string,
                role: data.get('role') as string,
                location: data.get('location') as string,
                link: data.get('link') as string,
                payMin: data.get('payMin') ? parseFloat(data.get('payMin') as string) : null,
                payMax: data.get('payMax') ? parseFloat(data.get('payMax') as string) : null,
                notes: data.get('notes') as string,
                appOpenDate: data.get('appOpenDate') as string,
                appCloseDate: data.get('appCloseDate') as string,
                internStartDate: data.get('internStartDate') as string,
                internEndDate: data.get('internEndDate') as string

            })
            .where(eq(applications.id, appId));

        return { success: true };
    },

    // Action to add a new event (Status Change)
    addEvent: async ({ request, params }) => {
        const appId = parseInt(params.id);
        const data = await request.formData();

        await db.insert(applicationEvents).values({
            applicationId: appId,
            eventType: data.get('eventType') as string,
            eventDate: data.get('eventDate') as string,
            details: data.get('details') as string,
        });

        return { success: true };
    },

    deleteApp: async ({ params }) => {
        const appId = parseInt(params.id);
        await db.delete(applicationEvents).where(eq(applicationEvents.applicationId, appId));
        await db.delete(applications).where(eq(applications.id, appId));
        throw redirect(303, '/');
    },

    deleteEvent: async ({ request }) => {
        const data = await request.formData();
        const eventId = parseInt(data.get('eventId') as string);

        if (!eventId) return { success: false };

        await db.delete(applicationEvents)
            .where(eq(applicationEvents.id, eventId));

        return { success: true };
    },
};