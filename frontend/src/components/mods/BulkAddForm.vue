<script setup lang="ts">
import {BULK_MAX_LINES, parseBulkLines} from "@/lib/bulk-add.ts"

const {starting} = defineProps<{
  starting: boolean
}>()

const emit = defineEmits<{
  submit: [text: string]
}>()

const text = defineModel<string>({default: ""})

const lineCount = computed(() => parseBulkLines(text.value).length)
const tooMany = computed(() => lineCount.value > BULK_MAX_LINES)
const hint = computed(() => {
  const mods = `${lineCount.value} ${lineCount.value === 1 ? "mod" : "mods"}`
  return `${mods}. Order matters: earlier lines win conflicts.`
})
</script>

<template>
  <v-form @submit.prevent="emit('submit', text)">
    <v-textarea
      v-model="text"
      label="Mods, one per line"
      placeholder="sodium&#10;Just Enough Items&#10;https://modrinth.com/mod/lithium"
      :hint="hint"
      :error-messages="tooMany ? `At most ${BULK_MAX_LINES} mods per list` : []"
      persistent-hint
      auto-grow
      rows="8"
      class="mb-4"
    />
    <v-alert
      type="info"
      variant="tonal"
      density="compact"
      class="mb-4"
      text="Each line is a mod name, slug or Modrinth/CurseForge link. Exact matches are added with their dependencies; everything else is left for you to review."
    />
    <div class="d-flex justify-end">
      <v-btn
        text="Add Mods"
        color="primary"
        type="submit"
        :loading="starting"
        :disabled="lineCount === 0 || tooMany"
      />
    </div>
  </v-form>
</template>
