<route lang="yaml">
meta:
  noAuth: true
</route>

<script setup lang="ts">

const router = useRouter()

const goHome = () => {
  router.push('/')
}
</script>

<template>
  <v-empty-state
    headline="Uh Oh!"
    title="404 Page not found"
    text="The page you are looking for does not seem to exist."
    icon="mdi-magnify"
    color="primary"
    action-text="Return Home"
    @click:action="goHome"
  />
</template>

