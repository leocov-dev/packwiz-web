<script setup lang="ts">
import MinecraftVersion from "@/components/forms/MinecraftVersion.vue";
import Loader from "@/components/forms/Loader.vue";
import type {NewPackRequest} from "@/interfaces/requests.ts";
import {newPack} from "@/services/packs.service.ts";

const isVald = ref(false)
const loading = ref(false)

const data = ref({
  slug: '',
  name: '',
  packVersion: "0.1.0",
  description: '',
  minecraftVersion: "",
  loader: {
    name: undefined,
    version: '',
  },
  acceptableVersions: [],
})

const router = useRouter()

const buildRequest: () => NewPackRequest = () => {
  const form = data.value

  const nonVersion = ["Latest", "LatestSnapshot"]

  return {
    slug: form.slug,
    name: form.name,
    version: form.packVersion,
    description: form.description,
    minecraft: {
      version: nonVersion.includes(form.minecraftVersion || "") ? "" : form.minecraftVersion || "",
      latest: form.minecraftVersion === "Latest",
      snapshot: form.minecraftVersion === "Latest Snapshot",
    },
    loader: {
      name: (form.loader.name || "").toLowerCase(),
      version: form.loader.version === "Latest" ? "" : form.loader.version,
      latest: form.loader.version === "Latest",
    },
    acceptableVersions: form.acceptableVersions,
  }
}

const submitForm = async () => {
  loading.value = true
  const request = buildRequest()

  await newPack(request)

  loading.value = false

  await router.push({path: `/packs/${request.slug}`})
}

</script>

<template>
  <div
    class="ma-6"
  >
    <v-card>
      <v-card-title class="d-flex align-center">
        <h1 class="me-5">
          New Pack
        </h1>
      </v-card-title>

      <v-form
        v-model="isVald"
        validate-on="eager"
        class="ma-6"
        @submit.prevent="submitForm"
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

        <v-row>
          <v-col>
            <v-divider />
          </v-col>
        </v-row>

        <v-row>
          <v-col>
            <MinecraftVersion
              v-model:version="data.minecraftVersion"
              :include-latest="true"
            />
          </v-col>
          <v-col>
            <AcceptableVersions
              v-model:acceptable-versions="data.acceptableVersions"
            />
          </v-col>
        </v-row>

        <v-row>
          <v-col>
            <v-divider />
          </v-col>
        </v-row>

        <v-row>
          <v-col>
            <Loader
              v-model:loader="data.loader.name"
              v-model:version="data.loader.version"
            />
          </v-col>
        </v-row>

        <v-row>
          <v-col>
            <v-divider />
          </v-col>
        </v-row>

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
            :disabled="loading"
            text="Cancel"
            class="me-6"
            to="/packs"
          />
          <v-btn
            :loading="loading"
            :disabled="loading || !isVald"
            text="Create"
            type="submit"
            color="primary"
          />
        </div>
      </v-form>
    </v-card>
  </div>
</template>

<style scoped lang="sass">

</style>
