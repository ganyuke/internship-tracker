CREATE TABLE `application_events` (
	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	`application_id` integer NOT NULL,
	`event_type` text NOT NULL,
	`event_date` text NOT NULL,
	`details` text,
	FOREIGN KEY (`application_id`) REFERENCES `applications`(`id`) ON UPDATE no action ON DELETE no action
);
--> statement-breakpoint
CREATE TABLE `applications` (
	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	`company` text NOT NULL,
	`role` text NOT NULL,
	`location` text,
	`pay_min` real,
	`pay_max` real,
	`link` text,
	`notes` text,
	`app_open_date` text,
	`app_close_date` text,
	`intern_start_date` text,
	`intern_end_date` text
);
