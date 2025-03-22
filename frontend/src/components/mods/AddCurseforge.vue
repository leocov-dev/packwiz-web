<script setup lang="ts">


import {tr} from "vuetify/locale";

const {
  class: classData,
  packName,
} = defineProps({
  class: {
    type: String,
    default: "",
  },
  packName: {
    type: String,
    required: true,
  },
})

</script>

<template>
  <div :class="classData">
    <v-card>
      <v-card-title>
        <h1 class="me-5">
          {{ packName }}
        </h1>
      </v-card-title>

      <v-card-subtitle>
        <h3>Add New Curseforge Mod</h3>
      </v-card-subtitle>

      <v-card-text>
        <v-form>
          <v-text-field/>
        </v-form>
      </v-card-text>
    </v-card>
  </div>
</template>
