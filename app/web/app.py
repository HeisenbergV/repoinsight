import streamlit as st
import requests
from datetime import datetime
import json

# 配置页面
st.set_page_config(
    page_title="RepoInsight",
    page_icon="🔍",
    layout="wide"
)

# 设置API基础URL
API_BASE_URL = "http://localhost:8000/api/v1"

# 侧边栏导航
page = st.sidebar.radio(
    "导航",
    ("项目搜索", "热门项目", "已分析项目")
)

if page == "项目搜索":
    st.title("GitHub 项目智能分析平台 - 搜索")
    st.header("搜索项目")
    search_query = st.text_input("输入搜索关键词")
    if st.button("搜索"):
        if search_query:
            with st.spinner("正在搜索..."):
                response = requests.get(f"{API_BASE_URL}/repositories", params={"q": search_query})
                if response.status_code == 200:
                    repos = response.json()
                    for repo in repos:
                        with st.expander(f"{repo['full_name']} ({repo['stars']} ⭐)"):
                            st.write(f"**描述:** {repo['description']}")
                            st.write(f"**语言:** {repo['language']}")
                            st.write(f"**主题:** {repo['topics']}")
                            if repo.get('analysis') and repo['analysis'].get('content'):
                                st.write("**AI分析结果:**")
                                st.write(repo['analysis']['content'])
                            else:
                                if st.button("分析项目", key=repo['id']):
                                    with st.spinner("正在分析..."):
                                        analysis_response = requests.post(
                                            f"{API_BASE_URL}/analysis/analyze",
                                            json={"url": repo['url']}
                                        )
                                        if analysis_response.status_code == 200:
                                            st.success("分析完成！")
                                            st.write(analysis_response.json()['content'])
                                        else:
                                            st.error("分析失败")

elif page == "热门项目":
    st.title("GitHub 项目智能分析平台 - 热门项目")
    col1, col2 = st.columns(2)
    with col1:
        st.subheader("按星标数")
        response = requests.get(f"{API_BASE_URL}/repositories/top", params={"sort": "stars"})
        if response.status_code == 200:
            repos = response.json()
            for repo in repos:
                with st.expander(f"{repo['full_name']} ({repo['stars']} ⭐)"):
                    st.write(f"**描述:** {repo['description']}")
                    st.write(f"**语言:** {repo['language']}")
                    st.write(f"**主题:** {repo['topics']}")
                    if repo.get('analysis') and repo['analysis'].get('content'):
                        st.write("**AI分析结果:**")
                        st.write(repo['analysis']['content'])
    with col2:
        st.subheader("最近更新")
        response = requests.get(f"{API_BASE_URL}/repositories/top", params={"sort": "updated"})
        if response.status_code == 200:
            repos = response.json()
            for repo in repos:
                with st.expander(f"{repo['full_name']} ({repo['stars']} ⭐)"):
                    st.write(f"**描述:** {repo['description']}")
                    st.write(f"**语言:** {repo['language']}")
                    st.write(f"**主题:** {repo['topics']}")
                    if repo.get('analysis') and repo['analysis'].get('content'):
                        st.write("**AI分析结果:**")
                        st.write(repo['analysis']['content'])

elif page == "已分析项目":
    st.title("已分析项目")
    st.info("只展示有AI分析结果的项目")
    response = requests.get(f"{API_BASE_URL}/repositories")
    if response.status_code == 200:
        repos = response.json()
        analyzed_repos = [r for r in repos if r.get('analysis') and r['analysis'].get('content')]
        if not analyzed_repos:
            st.warning("暂无已分析项目")
        for repo in analyzed_repos:
            with st.expander(f"{repo['full_name']} ({repo['stars']} ⭐)"):
                st.write(f"**描述:** {repo['description']}")
                st.write(f"**语言:** {repo['language']}")
                st.write(f"**主题:** {repo['topics']}")
                st.write("**AI分析结果:**")
                st.write(repo['analysis']['content'])
    else:
        st.error("获取数据失败")

# 页脚
st.markdown("---")
st.markdown("RepoInsight - 智能GitHub项目分析平台") 