export interface Article {
	id: number;
	slug: string;
	title: string;
	content: string;
	description: string;
}

export interface IndexRouteProps {
	articles: Article[] | null;
}

export interface ArticleRouteProps {
	article: Article;
}
