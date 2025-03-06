<script setup lang="ts">
import {type Pack} from "@/interfaces/pack.ts";

const {pack} = defineProps<{ pack: Pack }>()

</script>

<template>
  <div
    class="ma-6"
  >
    <v-card>
      <v-card-title class="d-flex align-center">
        <div
          class="d-flex align-center me-auto"
        >
          <h1 class="me-5">
            {{ pack.title }}
          </h1>
        </div>
      </v-card-title>

      <v-card-subtitle>
        <h3>Add a new mod</h3>
      </v-card-subtitle>

      <v-card-text>
        <v-form>
          <v-text-field>Stuff</v-text-field>
        </v-form>
      </v-card-text>
    </v-card>
  </div>
</template>

<style scoped lang="sass">

</style>
