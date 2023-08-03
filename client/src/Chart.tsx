import type { Component } from 'solid-js';
import { onMount } from 'solid-js';
import { Chart, PolarAreaController, ArcElement, Tooltip, Legend, RadialLinearScale } from 'chart.js';
import type { IssuedCollectable } from './resources';
import { Rarity, rarityColor } from './resources';

export type DistributionChartProps = {
  collection: IssuedCollectable[];
}

Chart.register(PolarAreaController, Legend, Tooltip, ArcElement, RadialLinearScale)

let rarities = [
  Rarity.Common,
  Rarity.Uncommon,
  Rarity.Rare,
  Rarity.SuperRare,
  Rarity.UltraRare,
];

const statsForCollection = (collectables: IssuedCollectable[]) => {
  let res = new Map<Rarity, number>();

  for (const c of collectables) {
    let v = res.get(c.rarity) || 0;
    res.set(c.rarity, v + 1)
  }

  return res;
}

export const DistributionChart: Component<DistributionChartProps> = (props) => {

  let canvas: HTMLCanvasElement | undefined;

  let { collection } = props;

  let stats = statsForCollection(collection);

  let data = {
    labels: rarities,
    datasets: [
      {
        label: "Count",
        data: rarities.map(r => stats.get(r) || 0),
        backgroundColor: rarities.map(r => rarityColor(r)),
      }
    ]
  };

  onMount(() => {
    if (canvas === undefined) { return; }


    new Chart(canvas, {
      type: "polarArea",
      data: data,
      options: {
        responsive: true,
        maintainAspectRatio: false,
      },
    })

  })

  return (
    <canvas ref={canvas}></canvas>
  );

}

