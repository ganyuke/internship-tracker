<script lang="ts">
	import type { HomepageData } from '../types';

	export let data: HomepageData;

	const { apps, stats } = data;

	function payDisplay(min: number | null, max: number | null) {
		if (min != null && max != null)
			return { text: `$${min} - $${max}`, className: 'text-green-700' };
		if (min != null || max != null) return { text: `$${min ?? max}`, className: 'text-yellow-700' };
		return { text: '???', className: 'text-red-500' };
	}

  function statusDisplay(latestStatus: string | null) {
    if (latestStatus === null) {
      return "text-gray-500";
    }
    switch (latestStatus) {
      case "REJECTED":
        return "text-red-500";
      case "INTERVIEW":
        return "text-yellow-500";
      case "OFFER":
        return "text-purple-500";
      default:
        return "text-blue-600";
    }
	}
</script>

<div class="mx-auto max-w-6xl space-y-8 p-8">
	<header class="flex items-center justify-between">
		<h1 class="text-3xl font-bold">Internship Tracker</h1>
		<a href="/add" class="rounded-lg bg-blue-600 px-4 py-2 text-white hover:bg-blue-700"
			>Add Application</a
		>
	</header>

	<!-- Stats Grid -->
	<div class="grid grid-cols-4 gap-4">
		{#each Object.entries(stats) as [label, val]}
			<div class="rounded-xl border bg-white p-6 shadow-sm">
				<p class="text-sm font-semibold text-gray-500 uppercase">{label}</p>
				<p class="text-2xl font-bold">{val}</p>
			</div>
		{/each}
	</div>

	<!-- Calendar Link -->
	<div class="flex items-center justify-between rounded-lg border border-blue-100 bg-blue-50 p-4">
		<p class="text-sm text-blue-800">Sync interviews to your calendar:</p>
		<code class="rounded border border-blue-200 bg-white px-2 py-1 text-xs">/calendar.ics</code>
	</div>

	<!-- Table -->
	<div class="overflow-x-auto rounded-xl border">
		<table class="w-full bg-white text-left">
			<thead class="border-b bg-gray-50">
				<tr>
					<th class="p-4">Company</th>
					<th class="p-4">Role</th>
          <th class="p-4">Date Range</th>
					<th class="p-4">Location</th>
					<th class="p-4">Pay Range</th>
					<th class="p-4">Status</th>
				</tr>
			</thead>
			<tbody>
				{#each apps as app}
					{@const {pay, statusColor} = {pay: payDisplay(app.payMin, app.payMax), statusColor: statusDisplay(app.latestStatus)} }
					<tr
						class="cursor-pointer border-b transition hover:bg-gray-50"
						on:click={() => (window.location.href = `/app/${app.id}`)}
					>
						<td class="p-4 font-medium text-blue-600 hover:underline">
							{app.company}
						</td> <td class="p-4">{app.role}</td>
            <td class="p-4">{`${app.internStartDate ? app.internStartDate : "?"} - ${app.internEndDate ? app.internEndDate : "?"}`}</td>
						<td class="p-4 text-gray-600">{app.location || 'Remote'}</td>
						<td class={`p-4 font-mono text-sm ${pay.className}`}>
							{pay.text}
						</td>
						<td class={`p-4 text-xs font-bold tracking-wider ${statusColor} uppercase`}>{app.latestStatus ?? "Unknown"}</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
</div>
