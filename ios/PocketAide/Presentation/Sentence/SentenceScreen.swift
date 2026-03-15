// SentenceScreen.swift
// PocketAide

import SwiftUI

/// 문장 모음 목록을 표시하는 메인 화면.
///
/// AccessibilityIdentifier 목록:
/// - sentence_list_view                     : root container
/// - sentence_category_section_{name}       : 각 카테고리 섹션
/// - add_category_button                    : 카테고리 추가 버튼
/// - add_sentence_button                    : 문장 추가 버튼
/// - sentence_row_{content}                 : 각 문장 행
/// - sentence_edit_button_{content}         : 각 문장의 편집 버튼
struct SentenceScreen: View {

    @StateObject private var viewModel = SentenceViewModel()

    var body: some View {
        NavigationStack {
            Group {
                if viewModel.categories.isEmpty && !viewModel.isLoading {
                    emptyView
                } else {
                    sentenceListView
                }
            }
            .navigationTitle("문장 모음")
            .navigationBarTitleDisplayMode(.large)
            .toolbar {
                ToolbarItem(placement: .navigationBarTrailing) {
                    HStack {
                        Button {
                            viewModel.showAddCategorySheet = true
                        } label: {
                            Image(systemName: "folder.badge.plus")
                        }
                        .accessibilityIdentifier("add_category_button")

                        Button {
                            viewModel.showAddSentenceSheet = true
                        } label: {
                            Image(systemName: "plus")
                        }
                        .accessibilityIdentifier("add_sentence_button")
                    }
                }
            }
            .sheet(isPresented: $viewModel.showAddCategorySheet) {
                CategoryEditModal(viewModel: viewModel)
            }
            .sheet(isPresented: $viewModel.showAddSentenceSheet) {
                SentenceEditModal(viewModel: viewModel)
            }
            .sheet(item: $viewModel.editingSentence) { sentence in
                SentenceEditSheet(viewModel: viewModel, sentence: sentence)
            }
            .task {
                await viewModel.loadData()
            }
            .refreshable {
                await viewModel.loadData()
            }
        }
        .accessibilityIdentifier("sentence_list_view")
    }

    // MARK: - Subviews

    private var sentenceListView: some View {
        List {
            ForEach(viewModel.categories) { category in
                Section {
                    let categorySentences = viewModel.sentences(for: category)
                    if categorySentences.isEmpty {
                        Text("문장이 없습니다")
                            .foregroundStyle(.secondary)
                            .font(.subheadline)
                    } else {
                        ForEach(categorySentences) { sentence in
                            SentenceRow(sentence: sentence, viewModel: viewModel)
                        }
                    }
                } header: {
                    Text(category.name)
                }
                .accessibilityIdentifier("sentence_category_section_\(category.name)")
            }
        }
    }

    private var emptyView: some View {
        VStack(spacing: 16) {
            Image(systemName: "text.quote")
                .font(.system(size: 60))
                .foregroundStyle(.secondary)
            Text("문장 모음이 없습니다")
                .font(.headline)
            Text("카테고리를 추가하고 문장을 모아보세요")
                .font(.subheadline)
                .foregroundStyle(.secondary)
        }
        .frame(maxWidth: .infinity, maxHeight: .infinity)
    }
}

// MARK: - SentenceRow

/// 문장 목록의 개별 행 뷰.
private struct SentenceRow: View {

    let sentence: Sentence
    let viewModel: SentenceViewModel

    var body: some View {
        HStack {
            Text(sentence.content)
                .font(.body)

            Spacer()

            Button {
                viewModel.editingSentence = sentence
            } label: {
                Image(systemName: "pencil")
                    .foregroundStyle(.secondary)
            }
            .buttonStyle(.plain)
            .accessibilityIdentifier("sentence_edit_button_\(sentence.content)")
        }
        .accessibilityIdentifier("sentence_row_\(sentence.content)")
        .contextMenu {
            Button(role: .destructive) {
                Task {
                    await viewModel.deleteSentence(id: sentence.id)
                }
            } label: {
                Label("삭제", systemImage: "trash")
            }
        }
    }
}
