<script setup lang="ts">

import {getCookieBool, setCookieBool} from "@/services/cookies.service";

const cookiePolicyName = 'cookiePolicyAccepted'

const showCookieNotice = ref(true)

const acceptCookiePolicy = () => {
  showCookieNotice.value = false
  setCookieBool(cookiePolicyName, true, 180)
}

onMounted(() => {
  showCookieNotice.value = !getCookieBool(cookiePolicyName)
})

</script>

<template>
  <v-bottom-sheet v-model="showCookieNotice">
    <v-card class="pa-5">
      <v-card-title>Cookies Policy Notice</v-card-title>
      <v-card-text>
        We use essential first-party cookies to ensure proper functioning of the website,
        such as maintaining your login session or preferences. These cookies are necessary for
        the site to function and do not collect personal data. No third-party cookies are
        used.
      </v-card-text>

      <v-card-actions>
        <v-btn
          variant="tonal"
          @click="acceptCookiePolicy"
        >
          Accept
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-bottom-sheet>
</template>


