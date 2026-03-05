import { sqliteTable, text, integer, real } from 'drizzle-orm/sqlite-core';

export const applications = sqliteTable('applications', {
	id: integer('id').primaryKey({ autoIncrement: true }),
	company: text('company').notNull(),
	role: text('role').notNull(),
	location: text('location'),
	payMin: real('pay_min'),
	payMax: real('pay_max'),
	link: text('link'),
	notes: text('notes'),
	appOpenDate: text('app_open_date'),
	appCloseDate: text('app_close_date'),
	internStartDate: text('intern_start_date'),
	internEndDate: text('intern_end_date'),
});

export const applicationEvents = sqliteTable('application_events', {
	id: integer('id').primaryKey({ autoIncrement: true }),
	applicationId: integer('application_id').references(() => applications.id).notNull(),
	eventType: text('event_type').notNull(),
	eventDate: text('event_date').notNull(),
	details: text('details'),
});

export type ApplicationEvent = typeof applicationEvents.$inferSelect;
export type Application = typeof applications.$inferSelect;
export type NewApplication = typeof applications.$inferInsert;