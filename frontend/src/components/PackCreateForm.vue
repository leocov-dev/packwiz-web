<script setup lang="ts">

import MinecraftVersion from "@/components/forms/MinecraftVersion.vue";
import Loader from "@/components/forms/Loader.vue";

const data = ref({
  slug: '',
  name: '',
  packVersion: "1.0.0",
  description: '',
  minecraftVersion: undefined,
  loader: {
    name: undefined,
    version: '',
  },
  acceptableVersions: [],
})

</script>

<template>
  <div
    class="ma-6"
  >
    <v-card>

      <v-form
        class="ma-6"
      >
        <v-row>
          <v-col class="v-col-md-9 v-col-sm-8">
            <SlugAndName
              v-model:slug="data.slug"
              v-model:name="data.name"
            />
          </v-col>
          <v-col class="v-col-md-3 v-col-sm-4">
            <v-text-field
              v-model="data.packVersion"
              label="Pack Version"
            />
          </v-col>
        </v-row>

        <v-row><v-col><v-divider /></v-col></v-row>

        <v-row>
          <v-col>
            <MinecraftVersion
              v-model:version="data.minecraftVersion"
            />
          </v-col>
          <v-col>
            <AcceptableVersions
              v-model:acceptable-versions="data.acceptableVersions"
            />
          </v-col>
        </v-row>

        <v-row><v-col><v-divider /></v-col></v-row>

        <v-row>
          <v-col>
            <Loader
              v-model:loader="data.loader.name"
              v-model:version="data.loader.version"
            />
          </v-col>
        </v-row>

        <v-row><v-col><v-divider /></v-col></v-row>

        <v-row>
          <v-col>
            <v-textarea
              v-model="data.description"
              label="Description"
            />
          </v-col>
        </v-row>

        <div class="d-flex justify-end">
          <v-btn
            text="Cancel"
            class="me-6"
            to="/packs"
          />
          <v-btn
            text="Create"
          />
        </div>
      </v-form>
    </v-card>
  </div>
</template>

<style scoped lang="sass">

</style>
