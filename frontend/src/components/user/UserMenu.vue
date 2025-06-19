<script setup lang="ts">
import {useAuthStore} from "@/stores/auth.ts";

const authStore = useAuthStore()

const {width = 350} = defineProps<{ width?: number }>()
const showMenu = ref(false)

</script>

<template>
  <v-menu
    v-if="authStore.user"
    v-model="showMenu"
    location="top start"
    origin="top start"
    transition="scale-transition"
  >
    <template #activator="{ props }">
      <v-btn
        v-bind="props"
        link
        rounded="pill"
        size="large"
        class="me-5"
        variant="flat"
        color="secondary"
        elevation="4"
      >
        <v-avatar
          class="ms-n5"
          start
          :color="authStore.user.isAdmin ? 'warning' : 'primary'"
        >
          <v-icon icon="mdi-account" />
        </v-avatar>
        <v-responsive
          class="text-truncate"
          :max-width="width"
        >
          {{ authStore.user.username }}
        </v-responsive>
      </v-btn>
    </template>

    <v-card :width="width">
      <v-list bg-color="primary-darken-2">
        <v-list-item>
          <template #prepend>
            <v-avatar
              start
              :color="authStore.user.isAdmin ? 'warning' : 'primary'"
            >
              <v-icon icon="mdi-account" />
            </v-avatar>
          </template>

          <v-list-item-title>{{ authStore.user.username }}</v-list-item-title>

          <!--          <v-list-item-subtitle>{{ authStore.user.email }}</v-list-item-subtitle>-->
        </v-list-item>
      </v-list>

      <v-list>
        <v-list-item
          prepend-icon="mdi-account-edit"
          link
          to="/user/profile"
        >
          <v-list-item-subtitle>User Profile</v-list-item-subtitle>
        </v-list-item>

        <v-list-item
          prepend-icon="mdi-theme-light-dark"
        >
          <ThemeSwitch />
        </v-list-item>

        <v-list-item
          prepend-icon="mdi-logout"
          link
          to="/auth/logout"
        >
          <v-list-item-subtitle>Log Out</v-list-item-subtitle>
        </v-list-item>
      </v-list>
    </v-card>
  </v-menu>
</template>
