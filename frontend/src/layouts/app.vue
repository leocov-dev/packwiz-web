<script lang="ts" setup>
import {ref} from 'vue'
import UserMenu from "@/components/UserMenu.vue";

const rail = ref(false)

const userPillWidth = 100


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
      class="d-flex"
    >
      <v-app-bar-nav-icon
        icon="mdi-layers"
        to="/"
        v-ripple="false"
      ></v-app-bar-nav-icon>
      <v-app-bar-title
        link text="Packwiz Web"
        class="ms-2"
      ></v-app-bar-title>

      <v-container
        class="pa-0 mr-2"
        :max-width="userPillWidth"
      >
        <UserMenu :width="userPillWidth"/>
      </v-container>
    </v-app-bar>

    <v-navigation-drawer
      app
      permanent
      expand-on-hover
      :rail="rail"
    >
      <Navigation/>
    </v-navigation-drawer>

    <v-main class="overflow-y-scroll">
      <router-view class="mb-10"/>
    </v-main>

    <CookiesWarn/>
  </v-app>
</template>
