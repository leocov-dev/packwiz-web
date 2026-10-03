<script setup lang="ts">
import {
  CategoryScale,
  Chart as ChartJS,
  Filler,
  LinearScale,
  LineElement,
  PointElement,
  Tooltip,
} from "chart.js";
import {Line} from "vue-chartjs";
import {useTheme} from "vuetify";
import type {AccessDay} from "@/interfaces/access.ts";
import {formatDay, seriesValues, withAlpha, type AccessKind} from "@/lib/access.ts";

ChartJS.register(CategoryScale, LinearScale, LineElement, PointElement, Filler, Tooltip)

type SeriesSpec = { kind: AccessKind, label: string, color: 'primary' | 'success' | 'error' }

const {series, lines, compact = false, height = 220, width = undefined} = defineProps<{
  series: readonly AccessDay[]
  lines: SeriesSpec[]
  compact?: boolean
  height?: number
  width?: number
}>()

const theme = useTheme()

const chartData = computed(() => {
  const colors = theme.current.value.colors
  return {
    labels: series.map((day) => formatDay(day.date)),
    datasets: lines.map((line) => ({
      label: line.label,
      data: seriesValues(series, line.kind),
      borderColor: colors[line.color],
      backgroundColor: withAlpha(colors[line.color] ?? '', 0.15),
      fill: true,
      tension: 0.3,
      borderWidth: 2,
      pointRadius: compact ? 0 : 3,
      pointHoverRadius: 4,
    })),
  }
})

const chartOptions = computed(() => {
  const colors = theme.current.value.colors
  const text = withAlpha(colors['on-surface'] ?? '', 0.6)
  const grid = withAlpha(colors['on-surface'] ?? '', 0.1)
  return {
    responsive: true,
    maintainAspectRatio: false,
    interaction: {mode: 'index' as const, intersect: false},
    plugins: {
      legend: {display: false},
      tooltip: {enabled: !compact},
    },
    scales: {
      x: {
        display: !compact,
        grid: {display: false},
        ticks: {color: text, maxTicksLimit: 8, maxRotation: 0},
      },
      y: {
        display: !compact,
        beginAtZero: true,
        grid: {color: grid},
        border: {display: false},
        ticks: {color: text, precision: 0, maxTicksLimit: 5},
      },
    },
  }
})
</script>

<template>
  <v-sheet
    class="position-relative"
    color="transparent"
    :height="height"
    :width="width"
  >
    <Line
      :data="chartData"
      :options="chartOptions"
    />
  </v-sheet>
</template>
