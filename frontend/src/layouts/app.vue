<script lang="ts" setup>
import {ref} from 'vue'
import UserMenu from "@/components/user/UserMenu.vue";

const rail = ref(false)

const updateRail = () => {
  rail.value = window.innerWidth < 800
}

// Set initial value
onMounted(() => {
  updateRail()
  window.addEventListener('resize', updateRail)
})

onUnmounted(() => {
  window.removeEventListener('resize', updateRail)
})

</script>


<template>
  <v-app>
    <v-app-bar
      app
      class="d-flex justify-start"
      color="primary"
      elevation="2"
    >
      <v-btn
        v-ripple="false"
        icon
        to="/"
      >
        <v-img
          src="/favicon-32x32.png"
          width="32"
          height="32"
        />
      </v-btn>
      <v-app-bar-title
        link
        class="ms-2 flex-grow-0"
      >
        Packwiz Web
      </v-app-bar-title>

      <v-spacer />
      <UserMenu />
    </v-app-bar>

    <v-navigation-drawer
      app
      permanent
      expand-on-hover
      :rail="rail"
    >
      <Navigation />
    </v-navigation-drawer>

    <v-main class="overflow-y-scroll">
      <router-view class="mb-10" />
    </v-main>

    <CookiesWarn />
  </v-app>
</template>
