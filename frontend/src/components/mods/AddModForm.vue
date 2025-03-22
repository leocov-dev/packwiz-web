<script setup lang="ts">
import {type Pack} from "@/interfaces/pack.ts";

const {pack} = defineProps<{ pack: Pack }>()

</script>

<template>
  <AddCurseforge
    :pack-name="pack.packData?.name || pack.slug"
  />
</template>

<style scoped lang="sass">

</style>
