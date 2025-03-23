<script setup lang="ts">
const router = useRouter()

const goHome = () => {
  router.push('/')
}
</script>

<template>
  <v-empty-state
    headline="Developer at Work!"
    title="This area is under construction, your patience is appreciated."
    icon="mdi-account-hard-hat"
    action-text="Return Home"
    color="warning"
    @click:action="goHome"
  ></v-empty-state>
</template>

