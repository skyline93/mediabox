<template>
    <div class="album-details">
        <div class="header">
            <h1>Album Photos</h1>
            <button class="upload-button" @click="openModal">上传图片</button>
        </div>

        <div v-if="isModalOpen" class="modal">
            <div class="modal-content">
                <span class="close" @click="closeModal">&times;</span>
                <div class="modal-header">
                    <h5>上传图片</h5>
                </div>
                <div class="modal-body">
                    <input type="file" ref="fileInput" @change="handleFileChange" multiple />
                </div>
                <div class="modal-footer">
                    <button @click="confirmUpload">确定</button>
                    <button @click="closeModal">取消</button>
                </div>
            </div>
        </div>

        <div v-if="loading" class="loading">加载中...</div>
        <div v-else class="gallery-container">
            <div ref="lightGalleryRef" class="gallery">
                <LightGallery :speed="500" licenseKey="0000-0000-000-0000">
                    <a v-for="(image, index) in images" :key="index" :href="imageUrls[index]" class="gallery-item">
                        <img :src="imageUrls[index]" :alt="image.file_name" />
                    </a>
                </LightGallery>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref, onMounted, nextTick } from 'vue';
import axiosInstance from '../services/axiosInstance';
import { BASE_URL } from '../config';
import { useRoute } from 'vue-router';
import lightGallery from 'lightgallery';
import LightGallery from 'lightgallery/vue';
import 'lightgallery/css/lightgallery.css';

const images = ref([]);
const imageUrls = ref([]);
const loading = ref(true);
const lightGalleryRef = ref(null);
const isModalOpen = ref(false);
const selectedFiles = ref([]);
const route = useRoute();
const albumId = route.params.id;

const fetchImages = async () => {
    try {
        const response = await axiosInstance.get(`/api/v1/photo?album_id=${albumId}`);
        images.value = response.data.data;
        imageUrls.value = await Promise.all(response.data.data.map(async (image) => {
            const url = getFullImageUrl(image.link);
            return await getImageBlob(url);
        }));
    } catch (error) {
        console.error('Error fetching images:', error);
    } finally {
        loading.value = false;
    }
};

const getFullImageUrl = (path) => {
    if (typeof path !== 'string') {
        console.error('Invalid path:', path);
        return '';
    }

    if (path.startsWith('http')) {
        return path;
    } else {
        try {
            const trimmedPath = path.replace(/^\//, '');
            const url = new URL(trimmedPath, BASE_URL);
            return url.href;
        } catch (error) {
            console.error('Invalid URL:', path, error);
            return '';
        }
    }
};

const getImageBlob = async (url) => {
    try {
        const response = await axiosInstance.get(url, { responseType: 'blob' });
        return URL.createObjectURL(response.data);
    } catch (error) {
        console.error('Error fetching image:', error);
        return '';
    }
};

const openModal = () => {
    isModalOpen.value = true;
};

const closeModal = () => {
    isModalOpen.value = false;
    selectedFiles.value = [];
};

const handleFileChange = (event) => {
    selectedFiles.value = Array.from(event.target.files);
};

const confirmUpload = async () => {
    if (selectedFiles.value.length === 0) {
        alert('请选择至少一个文件');
        return;
    }

    const formData = new FormData();
    selectedFiles.value.forEach(file => {
        formData.append('files[]', file);
    });
    formData.append('album_id', albumId);

    try {
        await axiosInstance.post('/api/v1/photo/upload', formData, {
            headers: {
                'Content-Type': 'multipart/form-data',
            },
        });
        fetchImages();
        closeModal();
    } catch (error) {
        console.error('Error uploading images:', error);
    }
};

onMounted(async () => {
    const lgThumbnail = await import('lightgallery/plugins/thumbnail/lg-thumbnail.es5.js');
    const lgFullscreen = await import('lightgallery/plugins/fullscreen/lg-fullscreen.es5.js');

    nextTick(() => {
        lightGallery(lightGalleryRef.value, {
            plugins: [lgThumbnail.default, lgFullscreen.default],
            thumbnail: true,
            fullscreen: true,
        });
    });

    fetchImages();
});
</script>


<style scoped>
.header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 10px;
    padding: 0 10px;
    box-sizing: border-box;
}

.upload-button {
    background-color: #42b983;
    color: white;
    border: none;
    padding: 10px 20px;
    border-radius: 5px;
    cursor: pointer;
    transition: background-color 0.3s ease;
    margin-left: auto;
}

.upload-button:hover {
    background-color: #369b72;
}

.modal {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 1000;
}

.modal-content {
    background: white;
    padding: 20px;
    border-radius: 8px;
    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2);
    max-width: 500px;
    width: 100%;
    box-sizing: border-box;
    position: relative;
}

.close {
    position: absolute;
    top: 10px;
    right: 10px;
    cursor: pointer;
    font-size: 24px;
    color: #333;
}

.close:hover {
    color: #42b983;
}

.modal-header {
    font-size: 1.5em;
    margin-bottom: 15px;
}

.modal-body {
    margin-bottom: 20px;
}

.modal-footer {
    text-align: right;
}

.modal-footer button {
    background-color: #42b983;
    color: white;
    border: none;
    padding: 10px 20px;
    border-radius: 5px;
    cursor: pointer;
    transition: background-color 0.3s ease;
    margin-left: 10px;
}

.modal-footer button:hover {
    background-color: #369b72;
}

.album-details {
    position: relative;
}

.gallery-container {
    padding: 10px;
    box-sizing: border-box;
}

.gallery {
    display: grid;
    grid-template-columns: auto-fill;
    gap: 5px;
}

.gallery-item {
    width: 100%;
}

.gallery-item img {
    width: 15%;
    height: 5%;
    object-fit: cover;
}

.loading {
    font-size: 1.2em;
    text-align: center;
    width: 100%;
}
</style>
