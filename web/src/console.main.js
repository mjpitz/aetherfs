/**
 * Copyright (C) The AetherFS Authors - All Rights Reserved
 * See LICENSE for more information.
 */

import {createApp} from 'vue'
import {createRouter, createWebHistory} from 'vue-router'

import Console from '@/components/Console.vue'
import DatasetDetails from '@/components/views/DatasetDetails.vue'
import DatasetList from '@/components/views/DatasetList.vue'
import DatasetOverview from '@/components/views/DatasetOverview.vue'
import TagDetails from '@/components/views/TagDetails.vue'

const router = createRouter({
    history: createWebHistory(process.env.BASE_URL),
    routes: [
        {path: "/", name: "DatasetList", component: DatasetList},
        {
            path: "/dataset/:dataset", name: "DatasetDetails", component: DatasetDetails,
            children: [
                {
                    path: '',
                    name: 'DatasetOverview',
                    component: DatasetOverview,
                },
                {
                    path: 'tag/:version',
                    name: 'TagDetails',
                    component: TagDetails,
                    children: [
                        {
                            path: 'tree/:path*',
                            name: 'FileTreeView',
                            component: TagDetails,
                        },
                    ],
                },
            ],
        },
        {
            path: "/dataset/:scope/:dataset", name: "ScopedDatasetDetails", component: DatasetDetails,
            children: [
                {
                    path: '',
                    name: 'ScopedDatasetOverview',
                    component: DatasetOverview,
                },
                {
                    path: 'tag/:version',
                    name: 'ScopedTagDetails',
                    component: TagDetails,
                    children: [
                        {
                            path: 'tree/:path*',
                            name: 'ScopedFileTreeView',
                            component: TagDetails,
                        },
                    ],
                },
            ],
        },
    ],
})

createApp(Console)
    .use(router)
    .mount('#app')
