import { db } from '$lib/server/db';
import { applications, applicationEvents, type Application } from '$lib/server/db/schema';
import { redirect } from '@sveltejs/kit';
import type { Actions } from './$types';

export const actions: Actions = {
    default: async ({ request }) => {
        const formData = await request.formData();

        // Extract values from the form
        const company = formData.get('company') as string;
        const role = formData.get('role') as string;
        const location = formData.get('location') as string;
        const link = formData.get('link') as string;
        const notes = formData.get('notes') as string;

        // Numbers come in as strings, so we need to parse them
        const payMin = formData.get('payMin') ? parseFloat(formData.get('payMin') as string) : null;
        const payMax = formData.get('payMax') ? parseFloat(formData.get('payMax') as string) : null;

        // Dates
        const appOpenDate = formData.get('appOpenDate') as string;
        const appCloseDate = formData.get('appCloseDate') as string;

        const internStartDate = formData.get('internStartDate') as string;
        const internEndDate = formData.get('internEndDate') as string;


        // Insert into Database
        const result = await db.insert(applications).values({
            company,
            role,
            location,
            link,
            notes,
            payMin,
            payMax,
            appOpenDate,
            appCloseDate,
            internStartDate,
            internEndDate
        });

        // Add the initial "APPLIED" event automatically
        await db.insert(applicationEvents).values({
            applicationId: result.lastInsertRowid as number,
            eventType: 'APPLIED',
            eventDate: new Date().toISOString().split('T')[0], // YYYY-MM-DD
            details: 'Initial application submitted'
        });

        // After saving, send the user back to the dashboard
        throw redirect(303, '/');
    }
};