<script lang="ts">
	let company = '';
	let role = '';
	let link = '';

	// Enhanced Autofill: Try to guess company AND role from URL
	function handleLinkBlur() {
		if (!link) return;

		// 1. Guess Company
		if (!company) {
			const companyMatch = link.match(/https?:\/\/(?:www\.)?([^\.]+)\./);
			if (companyMatch) {
				company = companyMatch[1].charAt(0).toUpperCase() + companyMatch[1].slice(1);
			}
		}

		// 2. Guess Role (e.g., /jobs/software-engineer-intern)
		if (!role) {
			const roleMatch = link.match(
				/([^\/]+)(?:-intern|software|engineer|designer|developer)[^\/]*$/i
			);
			if (roleMatch) {
				role = roleMatch[0]
					.split('-')
					.map((word) => word.charAt(0).toUpperCase() + word.slice(1))
					.join(' ');
			}
		}
	}
</script>

<div class="min-h-screen bg-gray-50 py-12">
	<form
		method="POST"
		class="mx-auto max-w-3xl space-y-6 rounded-2xl border border-gray-200 bg-white p-8 shadow-xl"
	>
		<div class="border-b border-gray-100 pb-4">
			<h2 class="text-2xl font-bold text-gray-800">New Application</h2>
			<p class="text-sm text-gray-500">Add a job listing to track (or ignore) later.</p>
		</div>

		<!-- Link & Company -->
		<div class="grid grid-cols-1 gap-6 md:grid-cols-2">
			<label class="block">
				<span class="text-sm font-semibold text-gray-700">Job Posting Link</span>
				<input
					name="link"
					bind:value={link}
					on:blur={handleLinkBlur}
					type="url"
					placeholder="https://..."
					class="mt-1 w-full rounded-lg border border-gray-300 p-2.5 transition outline-none focus:ring-2 focus:ring-black"
				/>
			</label>

			<label class="block">
				<span class="text-sm font-semibold text-gray-700">Company Name *</span>
				<input
					name="company"
					bind:value={company}
					required
					placeholder="e.g. Acme Corp"
					class="mt-1 w-full rounded-lg border border-gray-300 p-2.5 transition outline-none focus:ring-2 focus:ring-black"
				/>
			</label>
		</div>

		<!-- Role & Location -->
		<div class="grid grid-cols-1 gap-6 md:grid-cols-2">
			<label class="block">
				<span class="text-sm font-semibold text-gray-700">Job Role *</span>
				<input
					name="role"
					bind:value={role}
					required
					placeholder="e.g. Frontend Intern"
					class="mt-1 w-full rounded-lg border border-gray-300 p-2.5 transition outline-none focus:ring-2 focus:ring-black"
				/>
			</label>

			<label class="block">
				<span class="text-sm font-semibold text-gray-700">Location</span>
				<input
					name="location"
					placeholder="e.g. Remote / New York"
					class="mt-1 w-full rounded-lg border border-gray-300 p-2.5 transition outline-none focus:ring-2 focus:ring-black"
				/>
			</label>
		</div>

		<!-- Pay Range -->
		<div class="grid grid-cols-2 gap-6 rounded-xl bg-gray-50 p-4">
			<label class="block">
				<span class="text-sm font-semibold text-gray-700">Min Pay ($/hr)</span>
				<input
					type="number"
					step="0.01"
					name="payMin"
					placeholder="0.00"
					class="mt-1 w-full rounded-lg border border-gray-300 p-2.5 transition outline-none focus:ring-2 focus:ring-black"
				/>
			</label>
			<label class="block">
				<span class="text-sm font-semibold text-gray-700">Max Pay ($/hr)</span>
				<input
					type="number"
					step="0.01"
					name="payMax"
					placeholder="0.00"
					class="mt-1 w-full rounded-lg border border-gray-300 p-2.5 transition outline-none focus:ring-2 focus:ring-black"
				/>
			</label>
		</div>

		<!-- Dates -->
		<div class="grid grid-cols-2 gap-6">
			<label class="block">
				<span class="text-sm font-semibold text-gray-700">Application Opens</span>
				<input
					type="date"
					name="appOpenDate"
					class="mt-1 w-full rounded-lg border border-gray-300 p-2.5 transition outline-none focus:ring-2 focus:ring-black"
				/>
			</label>
			<label class="block">
				<span class="text-sm font-semibold text-gray-700">Application Closes</span>
				<input
					type="date"
					name="appCloseDate"
					class="mt-1 w-full rounded-lg border border-gray-300 p-2.5 transition outline-none focus:ring-2 focus:ring-black"
				/>
			</label>
            <label class="block">
				<span class="text-sm font-semibold text-gray-700">Start Date</span>
				<input
					type="date"
					name="internStartDate"
					class="mt-1 w-full rounded-lg border border-gray-300 p-2.5 transition outline-none focus:ring-2 focus:ring-black"
				/>
			</label>
			<label class="block">
				<span class="text-sm font-semibold text-gray-700">End Date</span>
				<input
					type="date"
					name="internEndDate"
					class="mt-1 w-full rounded-lg border border-gray-300 p-2.5 transition outline-none focus:ring-2 focus:ring-black"
				/>
			</label>
		</div>

		<!-- Notes -->
		<label class="block">
			<span class="text-sm font-semibold text-gray-700">Notes / Why am I procrastinating?</span>
			<textarea
				name="notes"
				rows="3"
				placeholder="e.g. Need to update my resume first..."
				class="mt-1 w-full rounded-lg border border-gray-300 p-2.5 transition outline-none focus:ring-2 focus:ring-black"
			></textarea>
		</label>

		<div class="flex gap-4 pt-4">
			<a
				href="/"
				class="flex-1 rounded-xl border border-gray-300 py-3 text-center font-semibold text-gray-600 transition hover:bg-gray-100"
			>
				Cancel
			</a>
			<button
				type="submit"
				class="flex-[2] rounded-xl bg-black py-3 font-bold text-white shadow-lg transition hover:bg-gray-800 active:scale-[0.98]"
			>
				Add Application
			</button>
		</div>
	</form>
</div>
