import { createRouter, createWebHistory } from 'vue-router';
import AlbumList from '../components/AlbumList.vue';
import AlbumDetail from '../components/AlbumDetail.vue';
import LoginForm from '../components/LoginForm.vue';
import LibraryPage from '../components/LibraryPage.vue';

const routes = [
    { path: '/login', name: 'LoginForm', component: LoginForm },
    { path: '/albums', name: 'AlbumList', component: AlbumList },
    { path: '/albums/:id', name: 'AlbumDetail', component: AlbumDetail },
    { path: '/library', name: 'LibraryPage', component: LibraryPage },
];

const router = createRouter({
    history: createWebHistory(),
    routes,
});

export default router;
