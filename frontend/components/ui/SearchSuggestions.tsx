

import React from "react";
import Tabs from "./Tabs";


const suggestions = [
	{
		title: "Post title heading will go here",
		description:
			"Lorem ipsum dolor sit amet, consectetur adipiscing elit. Suspendisse varius enim in eros.",
	},
	{
		title: "Post title heading will go here",
		description:
			"Lorem ipsum dolor sit amet, consectetur adipiscing elit. Suspendisse varius enim in eros.",
	},
	{
		title: "Post title heading will go here",
		description:
			"Lorem ipsum dolor sit amet, consectetur adipiscing elit. Suspendisse varius enim in eros.",
	},
];

function Suggestion({ title, description }: { title: string; description: string }) {
	return (
		<div
			className="
				flex
				w-[597px]
				h-[114px]
				p-6
				items-center
				rounded-none
				border-b
				border-parea-border
				bg-parea-white
				hover:bg-parea-yellow
				cursor-pointer
				transition-colors
				duration-200
			"
		>
			<div
				className="
					flex
					flex-col
					items-start
					gap-2
					flex-1
				"
			>
				<div
					className="
						self-stretch
						text-parea-black
						font-sans
						text-2xl
						font-bold
						leading-[1.4]
					"
				>
					{title}
				</div>
				<div
					className="
						self-stretch
						text-parea-black
						font-sans
						text-base
						font-normal
						leading-[1.5]
					"
				>
					{description}
				</div>
			</div>
		</div>
	);
}

export default function SearchSuggestions() {
	return (
		<div
			className="
				flex
				flex-col
				w-full
				max-w-full
				items-center
				border-b
				border-parea-black
				bg-parea-white
			"
		>
			<div
				className="
					flex
					flex-col
					min-h-[150px]
					px-8
					pt-4
					pb-8
					gap-4
					items-start
					self-stretch
					bg-parea-white
					rounded-xl
					shadow-lg
					absolute
					top-full
					left-0
					z-10
					w-full
				"
			>
				<div>
					<Tabs tabs={["Posts", "Events", "Users", "Groups"]} />
				</div>
				<div
					className="
						flex
						flex-col
						gap-4
						w-full
					"
				>
					{suggestions.map((s, i) => (
						<Suggestion key={i} title={s.title} description={s.description} />
					))}
				</div>
			</div>
		</div>
	);
}
