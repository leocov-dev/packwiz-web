<route lang="yaml">
meta:
  layout: app
</route>

<script lang="ts" setup>

import {useRoute} from "vue-router";
import PackDetails from "@/components/PackDetails.vue";

const route = useRoute<'/packs/[slug]'>()

const {slug} = route.params

</script>

<template>
  <PackDetails :slug="slug" />
</template>
