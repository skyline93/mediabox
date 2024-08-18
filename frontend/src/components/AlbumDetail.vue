<template>
    <div class="album-details">
        <div class="header">
            <h1>Album Photos</h1>
            <button class="upload-button" @click="openModal">上传图片</button>
            <button class="delete-selected-button" @click="deleteSelectedPhotos">删除选中的照片</button>
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
                        <input type="checkbox" v-model="selectedImages" :value="image.id" class="select-checkbox" @click.stop />
                    </a>
                </LightGallery>
            </div>
        </div>

        <div class="pagination">
            <button @click="previousPage" :disabled="page === 1">上一页</button>
            <span>当前页: {{ page }} / {{ totalPage }}</span>
            <button @click="nextPage" :disabled="page === totalPage">下一页</button>

            <label for="limit-select">每页显示数量：</label>
            <select id="limit-select" v-model="limit" @change="fetchImages" class="styled-select">
                <option value="5">5</option>
                <option value="10">10</option>
                <option value="20">20</option>
                <option value="50">50</option>
            </select>
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
const selectedImages = ref([]);
const route = useRoute();
const albumId = route.params.id;

const totalNum = ref(0);
const totalPage = ref(0);
const page = ref(1);
const limit = ref(10);

const fetchImages = async () => {
    loading.value = true;
    try {
        const response = await axiosInstance.get(`/api/v1/photo?album_id=${albumId}&page=${page.value}&limit=${limit.value}`);

        // 更新照片列表
        images.value = response.data.items;

        // 更新分页相关数据
        totalNum.value = response.data.total_num;
        totalPage.value = response.data.total_page;
        page.value = response.data.page;
        limit.value = response.data.limit;

        // 更新图片URL
        imageUrls.value = await Promise.all(response.data.items.map(async (image) => {
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

const deleteSelectedPhotos = async () => {
    if (selectedImages.value.length === 0) {
        alert('请选择要删除的照片');
        return;
    }

    try {
        await axiosInstance.delete('/api/v1/photo', { data: selectedImages.value });
        fetchImages();
        selectedImages.value = [];
    } catch (error) {
        console.error('Error deleting photos:', error);
    }
};

const previousPage = () => {
    if (page.value > 1) {
        page.value--;
        fetchImages();
    }
};

const nextPage = () => {
    if (page.value < totalPage.value) {
        page.value++;
        fetchImages();
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

.delete-selected-button {
    background-color: #e74c3c;
    color: white;
    border: none;
    padding: 10px 20px;
    border-radius: 5px;
    cursor: pointer;
    transition: background-color 0.3s ease;
    margin-left: 10px;
}

.delete-selected-button:hover {
    background-color: #c0392b;
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
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 10px;
    box-sizing: border-box;
}

.gallery {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
    gap: 5px;
    flex-direction: column;
    align-items: center;
}

.gallery-item {
    position: relative;
    width: auto;
    height: auto;
    max-width: 50px; 
    margin-bottom: 10px;
}

.gallery-item img {
    width: 100%;
    height: auto;
    object-fit: cover;
}

.select-checkbox {
    position: absolute;
    bottom: 10px;
    right: 10px;
    background: rgba(0, 0, 0, 0.5);
    border: none;
    cursor: pointer;
    z-index: 10;
}

.loading {
    font-size: 1.2em;
    text-align: center;
    width: 50%;
}

.pagination {
    display: flex;
    justify-content: center;
    align-items: center;
    margin-top: 20px;
}

.pagination button {
    background-color: #42b983;
    color: white;
    border: none;
    padding: 10px 20px;
    margin: 0 10px;
    border-radius: 5px;
    cursor: pointer;
}

.pagination button:disabled {
    background-color: #cccccc;
    cursor: not-allowed;
}

.upload-button,
.styled-select {
    background-color: #42b983;
    color: white;
    border: none;
    padding: 10px 20px;
    border-radius: 5px;
    cursor: pointer;
    transition: background-color 0.3s ease;
    margin-left: 10px;
}

.upload-button:hover,
.styled-select:hover {
    background-color: #369b72;
}

.styled-select {
    appearance: none;
    -webkit-appearance: none;
    -moz-appearance: none;
    text-align: center;
    padding-right: 30px;
    position: relative;
}

.select-limit {
    display: flex;
    align-items: center;
}

.select-limit label {
    margin-right: 10px;
}
</style>
