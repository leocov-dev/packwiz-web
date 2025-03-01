<script setup lang="ts">
import {useAuthStore} from "@/stores/auth";

const authStore = useAuthStore()

const {width = 100} = defineProps<{ width?: number }>()
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
      <v-chip
        v-bind="props"
        link
        pill
      >
        <v-avatar
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
      </v-chip>
    </template>

    <v-card width="300">
      <v-list bg-color="black">
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

          <!--              <v-list-item-subtitle>{{authStore.user.email}}</v-list-item-subtitle>-->
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

        <ThemeSwitch/>

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
