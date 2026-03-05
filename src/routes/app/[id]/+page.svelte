<script lang="ts">
	import type { PageData } from './$types';
	export let data: PageData;

	import { enhance } from '$app/forms'; // Add this import at the top
	import { eventTypes } from '../../../types';
	$: ({ app, events } = data); // Re-bind if data changes
</script>

<div class="mx-auto grid max-w-6xl grid-cols-1 gap-8 p-6 lg:grid-cols-3">
	<!-- Column 1 & 2: Edit Application Details -->
	<div class="space-y-6 lg:col-span-2">
		<div class="flex items-center justify-between">
			<a href="/" class="text-sm text-gray-500 hover:underline">← Back to Dashboard</a>
			<form
				method="POST"
				action="?/deleteApp"
				use:enhance={({ cancel }) => {
					if (!confirm('Are you sure you want to delete this application?')) {
						return cancel();
					}
				}}
			>
				<button class="rounded bg-red-50 px-4 py-2 text-sm text-red-500 hover:bg-red-200"
					>Delete App</button
				>
			</form>
		</div>

		<form
			method="POST"
			action="?/updateDetails"
			class="space-y-4 rounded-xl border bg-white p-6 shadow-sm"
		>
			<h2 class="text-xl font-bold">Edit Details</h2>
			<div class="grid grid-cols-2 gap-4">
				<label class="block">
					<span class="text-sm text-gray-500">Company</span>
					<input name="company" value={app.company} class="w-full rounded border p-2" />
				</label>
				<label class="block">
					<span class="text-sm text-gray-500">Role</span>
					<input name="role" value={app.role} class="w-full rounded border p-2" />
				</label>
			</div>
			<label class="block">
				<span class="text-sm text-gray-500">Link</span>
				<input name="link" value={app.link} class="w-full rounded border p-2" />
			</label>
			<div class="grid grid-cols-2 gap-4">
                <div class="grid grid-cols-2 gap-4">
                    <label class="block">
                        <span class="text-sm text-gray-500">Min Pay</span>
                        <input type="number" name="payMin" value={app.payMin} class="w-full rounded border p-2" />
                    </label>
                    <label class="block">
                        <span class="text-sm text-gray-500">Max Pay</span>
                        <input type="number" name="payMax" value={app.payMax} class="w-full rounded border p-2" />
                    </label>
                </div>
				<label class="block">
					<span class="text-sm text-gray-500">Location</span>
					<input name="location" value={app.location} class="w-full rounded border p-2" />
				</label>
			</div>
			<div class="grid grid-cols-4 gap-6">
				<label class="block">
					<span class="text-sm font-semibold text-gray-700">Application Opens</span>
					<input
						type="date"
						name="appOpenDate"
						value={app.appOpenDate}
						class="mt-1 w-full rounded-lg border border-gray-300 p-2.5 transition outline-none focus:ring-2 focus:ring-black"
					/>
				</label>
				<label class="block">
					<span class="text-sm font-semibold text-gray-700">Application Closes</span>
					<input
						type="date"
						name="appCloseDate"
						value={app.appCloseDate}
						class="mt-1 w-full rounded-lg border border-gray-300 p-2.5 transition outline-none focus:ring-2 focus:ring-black"
					/>
				</label>
				<label class="block">
					<span class="text-sm font-semibold text-gray-700">Start Date</span>
					<input
						type="date"
						name="internStartDate"
						value={app.internStartDate}
						class="mt-1 w-full rounded-lg border border-gray-300 p-2.5 transition outline-none focus:ring-2 focus:ring-black"
					/>
				</label>
				<label class="block">
					<span class="text-sm font-semibold text-gray-700">End Date</span>
					<input
						type="date"
						name="internEndDate"
						value={app.internEndDate}
						class="mt-1 w-full rounded-lg border border-gray-300 p-2.5 transition outline-none focus:ring-2 focus:ring-black"
					/>
				</label>
			</div>
			<label class="block">
				<span class="text-sm text-gray-500">Notes</span>
				<textarea name="notes" class="w-full rounded border p-2" rows="4"
					>{app.notes || ''}</textarea
				>
			</label>
			<button class="rounded-lg bg-black px-6 py-2 font-bold text-white">Save Changes</button>
		</form>
	</div>

	<!-- Column 3: Status Timeline -->
	<div class="space-y-6">
		<div class="rounded-xl border bg-white p-6 shadow-sm">
			<h2 class="mb-4 text-xl font-bold">Add Status Update</h2>
			<form method="POST" action="?/addEvent" class="space-y-3">
				<select name="eventType" class="w-full rounded border bg-gray-50 p-2" required>
					{#each eventTypes as type}
						<option value={type}>{type}</option>
					{/each}
				</select>
				<input
					type="date"
					name="eventDate"
					value={new Date().toISOString().split('T')[0]}
					class="w-full rounded border p-2"
					required
				/>
				<input
					type="text"
					name="details"
					placeholder="Additional details (e.g. 'with recruiter Sarah')"
					class="w-full rounded border p-2"
				/>
				<button class="w-full rounded-lg bg-blue-600 py-2 font-bold text-white hover:bg-blue-700"
					>Update Status</button
				>
			</form>
		</div>

		<div class="space-y-4">
			<h3 class="font-bold text-gray-700">Timeline</h3>
			<div class="relative ml-2 space-y-6 border-l-2 border-gray-200">
				{#each events as event}
					<div class="group relative ml-4">
						<!-- Added 'group' for hover effects -->
						<!-- The Dot -->
						<div
							class="absolute -left-[23px] mt-1.5 h-3 w-3 rounded-full border-2 border-white bg-blue-500"
						></div>

						<div class="flex items-start justify-between">
							<div>
								<p class="text-xs font-bold text-gray-500 uppercase">{event.eventDate}</p>
								<p class="font-bold text-gray-800">{event.eventType}</p>
								{#if event.details}
									<p class="text-sm text-gray-600">{event.details}</p>
								{/if}
							</div>

							<!-- DELETE EVENT BUTTON -->
							<form
								method="POST"
								action="?/deleteEvent"
								use:enhance={({ cancel }) => {
									if (!confirm('Are you sure you want to delete this event?')) {
										return cancel();
									}
								}}
							>
								<input type="hidden" name="eventId" value={event.id} />
								<button
									type="submit"
									class="p-1 text-gray-400 opacity-0 transition-opacity group-hover:opacity-100 hover:text-red-500"
									title="Delete event"
								>
									<svg
										xmlns="http://www.w3.org/2000/svg"
										class="h-4 w-4"
										fill="none"
										viewBox="0 0 24 24"
										stroke="currentColor"
									>
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											stroke-width="2"
											d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
										/>
									</svg>
								</button>
							</form>
						</div>
					</div>
				{/each}
			</div>
		</div>
	</div>
</div>
