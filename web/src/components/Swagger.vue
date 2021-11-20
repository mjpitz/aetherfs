<!--
 Copyright (C) The AetherFS Authors - All Rights Reserved
 See LICENSE for more information.
-->

<template>
  <div id="swagger-ui"/>
</template>

<script>
import { SwaggerUIBundle, SwaggerUIStandalonePreset } from "swagger-ui-dist"

export default {
  mounted() {
    SwaggerUIBundle({
      dom_id: "#swagger-ui",
      urls: [
        { name: "Agent API",   url: process.env.BASE_URL + "swagger/aetherfs/agent/v1/api.swagger.json" },
        { name: "Block API",   url: process.env.BASE_URL + "swagger/aetherfs/block/v1/api.swagger.json" },
        { name: "Dataset API", url: process.env.BASE_URL + "swagger/aetherfs/dataset/v1/api.swagger.json" },
      ],
      presets: [
        SwaggerUIBundle.presets.apis,
        SwaggerUIStandalonePreset,
      ],
      layout: "StandaloneLayout",
    })
  }
}
</script>

<style scoped>
@import "~normalize.css/normalize.css";
@import "~swagger-ui-dist/swagger-ui.css";


</style>
