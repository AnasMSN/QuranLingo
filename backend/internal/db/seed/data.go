package seed

// Word is a single Arabic vocabulary item with its English translation,
// the atomic unit of content for the course.
//
// Transliteration, Root and Occurrences are reference metadata carried over
// from the source material. They are not written to the database by the
// current seeder, but they are here so exercises can use them later without
// re-deriving the dataset.
type Word struct {
	Arabic          string
	English         string
	Transliteration string
	Root            string // Arabic trilateral root, empty for particles and proper nouns
	Type            string // "noun", "verb" or "particle"
	Occurrences     int    // times the word (with its combinations) occurs in the Qur'an
}

// Skill groups a themed set of words into one lesson.
type Skill struct {
	Code        string
	Title       string
	Description string
	Icon        string
	Words       []Word
}

const CourseCode = "quran-125"
const CourseTitle = "125 Words of the Qur'an"
const CourseDescription = "The 125 most frequent words of the Qur'an. Together they account for roughly 40,000 of the Qur'an's 78,000 words — about half of the text."

// Skills holds all 125 words in 25 themed lessons of 5 words each.
//
// Source: "125 Words of the Qur'an", compiled by Dr. Abdulazeez Abdulraheem,
// Understand Al-Qur'an Academy. Occurrence counts are the book's figures,
// counting the word together with its listed combinations.
var Skills = []Skill{
	{
		Code:        "names-of-allah",
		Title:       "Names of Allah",
		Description: "Allah and the divine names you meet in the very first verse.",
		Icon:        "☪️",
		Words: []Word{
			{Arabic: "اللَّه", English: "Allah", Transliteration: "Allah", Root: "", Type: "noun", Occurrences: 2550},
			{Arabic: "الرَّحْمٰن", English: "the Most Gracious", Transliteration: "Ar-Rahmaan", Root: "ر ح م", Type: "noun", Occurrences: 57},
			{Arabic: "الرَّحِيم", English: "the Most Merciful", Transliteration: "Ar-Raheem", Root: "ر ح م", Type: "noun", Occurrences: 116},
			{Arabic: "الْكَرِيم", English: "the Most Generous", Transliteration: "Al-Kareem", Root: "ك ر م", Type: "noun", Occurrences: 27},
			{Arabic: "الْعَظِيم", English: "the Great", Transliteration: "Al-'Azeem", Root: "ع ظ م", Type: "noun", Occurrences: 107},
		},
	},
	{
		Code:        "praise-and-lordship",
		Title:       "Praise & Lordship",
		Description: "The vocabulary of Al-Fatihah's opening praise.",
		Icon:        "🤲",
		Words: []Word{
			{Arabic: "الْحَمْد", English: "all praise and thanks", Transliteration: "Al-hamd", Root: "ح م د", Type: "noun", Occurrences: 149},
			{Arabic: "لِلَّهِ", English: "to Allah; for Allah", Transliteration: "lillaahi", Root: "", Type: "particle", Occurrences: 149},
			{Arabic: "رَبّ", English: "Lord", Transliteration: "Rabb", Root: "ر ب ب", Type: "noun", Occurrences: 971},
			{Arabic: "الْعَالَمِين", English: "the worlds", Transliteration: "Al-'aalameen", Root: "ع ل م", Type: "noun", Occurrences: 73},
			{Arabic: "سُبْحَان", English: "glory be to", Transliteration: "Subhaana", Root: "س ب ح", Type: "noun", Occurrences: 41},
		},
	},
	{
		Code:        "personal-pronouns",
		Title:       "Personal Pronouns",
		Description: "He, she, they, I, we.",
		Icon:        "🧍",
		Words: []Word{
			{Arabic: "هُوَ", English: "he", Transliteration: "huwa", Root: "", Type: "noun", Occurrences: 481},
			{Arabic: "هِيَ", English: "she; it", Transliteration: "hiya", Root: "", Type: "noun", Occurrences: 64},
			{Arabic: "هُمْ", English: "they", Transliteration: "hum", Root: "", Type: "noun", Occurrences: 444},
			{Arabic: "أَنَا", English: "I", Transliteration: "anaa", Root: "", Type: "noun", Occurrences: 68},
			{Arabic: "نَحْنُ", English: "we", Transliteration: "nahnu", Root: "", Type: "noun", Occurrences: 86},
		},
	},
	{
		Code:        "addressing-others",
		Title:       "Addressing Others",
		Description: "Speaking to someone: you, you all, and the call 'O!'",
		Icon:        "📣",
		Words: []Word{
			{Arabic: "أَنْتَ", English: "you (singular)", Transliteration: "anta", Root: "", Type: "noun", Occurrences: 81},
			{Arabic: "أَنْتُمْ", English: "you (plural)", Transliteration: "antum", Root: "", Type: "noun", Occurrences: 135},
			{Arabic: "إِيَّاكَ", English: "You alone", Transliteration: "iyyaaka", Root: "", Type: "noun", Occurrences: 24},
			{Arabic: "يَا", English: "O", Transliteration: "yaa", Root: "", Type: "particle", Occurrences: 361},
			{Arabic: "أَيُّهَا", English: "O you", Transliteration: "ayyuhaa", Root: "", Type: "particle", Occurrences: 153},
		},
	},
	{
		Code:        "relatives-and-questions",
		Title:       "Who & What",
		Description: "Relative pronouns and the words that ask.",
		Icon:        "❓",
		Words: []Word{
			{Arabic: "الَّذِي", English: "the one who", Transliteration: "alla'ee", Root: "", Type: "noun", Occurrences: 304},
			{Arabic: "الَّذِينَ", English: "those who", Transliteration: "alla'eena", Root: "", Type: "noun", Occurrences: 1080},
			{Arabic: "مَنْ", English: "who", Transliteration: "man", Root: "", Type: "noun", Occurrences: 831},
			{Arabic: "مَا", English: "that which", Transliteration: "maa", Root: "", Type: "noun", Occurrences: 2154},
			{Arabic: "مَاذَا", English: "what?", Transliteration: "maa'aa", Root: "", Type: "noun", Occurrences: 27},
		},
	},
	{
		Code:        "demonstratives",
		Title:       "This & That",
		Description: "Pointing at things near and far.",
		Icon:        "👉",
		Words: []Word{
			{Arabic: "هٰذَا", English: "this (masculine)", Transliteration: "haa'aa", Root: "", Type: "noun", Occurrences: 225},
			{Arabic: "هٰذِهِ", English: "this (feminine)", Transliteration: "haa'ihi", Root: "", Type: "noun", Occurrences: 47},
			{Arabic: "ذٰلِكَ", English: "that (masculine)", Transliteration: "'aalika", Root: "", Type: "noun", Occurrences: 478},
			{Arabic: "تِلْكَ", English: "that (feminine)", Transliteration: "tilka", Root: "", Type: "noun", Occurrences: 43},
			{Arabic: "هٰؤُلَاءِ", English: "these (people)", Transliteration: "haa'ulaa'i", Root: "", Type: "noun", Occurrences: 46},
		},
	},
	{
		Code:        "pointing-and-asking",
		Title:       "Those & Which",
		Description: "More pointers, plus the yes-or-no question particle.",
		Icon:        "🔎",
		Words: []Word{
			{Arabic: "أُولٰئِكَ", English: "those (people)", Transliteration: "ulaa'ika", Root: "", Type: "noun", Occurrences: 204},
			{Arabic: "هَلْ", English: "is? do? are?", Transliteration: "hal", Root: "", Type: "particle", Occurrences: 93},
			{Arabic: "أَيّ", English: "which", Transliteration: "ayy", Root: "", Type: "noun", Occurrences: 59},
			{Arabic: "أُولُو", English: "possessors of", Transliteration: "uloo", Root: "", Type: "noun", Occurrences: 43},
			{Arabic: "أَحَد", English: "one", Transliteration: "ahad", Root: "أ ح د", Type: "noun", Occurrences: 74},
		},
	},
	{
		Code:        "the-straight-path",
		Title:       "The Straight Path",
		Description: "Religion, the path, and what is good.",
		Icon:        "🛤️",
		Words: []Word{
			{Arabic: "الدِّين", English: "the religion; the judgment", Transliteration: "Ad-deen", Root: "د ي ن", Type: "noun", Occurrences: 92},
			{Arabic: "الصِّرَاط", English: "the path", Transliteration: "As-siraat", Root: "ص ر ط", Type: "noun", Occurrences: 45},
			{Arabic: "الْمُسْتَقِيم", English: "the straight", Transliteration: "Al-mustaqeem", Root: "ق و م", Type: "noun", Occurrences: 37},
			{Arabic: "خَيْر", English: "good; better", Transliteration: "khayr", Root: "خ ي ر", Type: "noun", Occurrences: 176},
			{Arabic: "أَحْسَن", English: "best", Transliteration: "ahsan", Root: "ح س ن", Type: "noun", Occurrences: 36},
		},
	},
	{
		Code:        "truth-and-misguidance",
		Title:       "Truth & Misguidance",
		Description: "The truth, the way, and those who lose it.",
		Icon:        "⚖️",
		Words: []Word{
			{Arabic: "الْحَقّ", English: "the truth", Transliteration: "Al-haqq", Root: "ح ق ق", Type: "noun", Occurrences: 247},
			{Arabic: "الضَّالِّين", English: "those who go astray", Transliteration: "Ad-daalleen", Root: "ض ل ل", Type: "noun", Occurrences: 14},
			{Arabic: "غَيْر", English: "other than", Transliteration: "ghayr", Root: "غ ي ر", Type: "noun", Occurrences: 147},
			{Arabic: "سَبِيل", English: "way", Transliteration: "sabeel", Root: "س ب ل", Type: "noun", Occurrences: 176},
			{Arabic: "نَصْر", English: "help; victory", Transliteration: "nasr", Root: "ن ص ر", Type: "noun", Occurrences: 80},
		},
	},
	{
		Code:        "believers-and-disbelievers",
		Title:       "Believers & Disbelievers",
		Description: "How the Qur'an names people by their response to it.",
		Icon:        "👥",
		Words: []Word{
			{Arabic: "مُسْلِم", English: "Muslim", Transliteration: "muslim", Root: "س ل م", Type: "noun", Occurrences: 42},
			{Arabic: "مُؤْمِن", English: "believer", Transliteration: "mu'min", Root: "أ م ن", Type: "noun", Occurrences: 230},
			{Arabic: "كَافِر", English: "disbeliever", Transliteration: "kaafir", Root: "ك ف ر", Type: "noun", Occurrences: 134},
			{Arabic: "مُشْرِك", English: "polytheist", Transliteration: "mushrik", Root: "ش ر ك", Type: "noun", Occurrences: 49},
			{Arabic: "صَالِح", English: "righteous person", Transliteration: "saalih", Root: "ص ل ح", Type: "noun", Occurrences: 136},
		},
	},
	{
		Code:        "mankind",
		Title:       "Mankind",
		Description: "People, nations, and the soul within.",
		Icon:        "🧑‍🤝‍🧑",
		Words: []Word{
			{Arabic: "الْإِنْسَان", English: "mankind", Transliteration: "Al-insaan", Root: "أ ن س", Type: "noun", Occurrences: 65},
			{Arabic: "النَّاس", English: "the people", Transliteration: "an-naas", Root: "ن و س", Type: "noun", Occurrences: 241},
			{Arabic: "قَوْم", English: "a nation", Transliteration: "qawm", Root: "ق و م", Type: "noun", Occurrences: 383},
			{Arabic: "نَفْس", English: "soul; self", Transliteration: "nafs", Root: "ن ف س", Type: "noun", Occurrences: 293},
			{Arabic: "عِبَاد", English: "servants", Transliteration: "'ibaad", Root: "ع ب د", Type: "noun", Occurrences: 125},
		},
	},
	{
		Code:        "the-unseen",
		Title:       "The Unseen",
		Description: "Angels, jinn, Satan, and false partners.",
		Icon:        "🌌",
		Words: []Word{
			{Arabic: "الشَّيْطَان", English: "Satan", Transliteration: "Ash-shaytaan", Root: "ش ط ن", Type: "noun", Occurrences: 88},
			{Arabic: "مَلَك", English: "angel", Transliteration: "malak", Root: "م ل ك", Type: "noun", Occurrences: 88},
			{Arabic: "جِنَّة", English: "jinn", Transliteration: "jinnah", Root: "ج ن ن", Type: "noun", Occurrences: 32},
			{Arabic: "شَرِيك", English: "partner", Transliteration: "shareek", Root: "ش ر ك", Type: "noun", Occurrences: 40},
			{Arabic: "إِلٰه", English: "god", Transliteration: "ilaah", Root: "أ ل ه", Type: "noun", Occurrences: 145},
		},
	},
	{
		Code:        "prophets-and-messengers",
		Title:       "Prophets & Messengers",
		Description: "Those who carried the message, and the command to speak.",
		Icon:        "🕌",
		Words: []Word{
			{Arabic: "مُحَمَّد", English: "Muhammad", Transliteration: "Muhammad", Root: "ح م د", Type: "noun", Occurrences: 4},
			{Arabic: "رَسُول", English: "messenger", Transliteration: "rasool", Root: "ر س ل", Type: "noun", Occurrences: 332},
			{Arabic: "النَّبِيّ", English: "the Prophet", Transliteration: "an-nabiyy", Root: "ن ب أ", Type: "noun", Occurrences: 75},
			{Arabic: "إِبْرَاهِيم", English: "Ibrahim", Transliteration: "Ibraaheem", Root: "", Type: "noun", Occurrences: 69},
			{Arabic: "قُلْ", English: "say!", Transliteration: "qul", Root: "ق و ل", Type: "verb", Occurrences: 332},
		},
	},
	{
		Code:        "the-book",
		Title:       "The Book",
		Description: "Scripture, revelation, and its verses.",
		Icon:        "📖",
		Words: []Word{
			{Arabic: "كِتَاب", English: "a book", Transliteration: "kitaab", Root: "ك ت ب", Type: "noun", Occurrences: 261},
			{Arabic: "الْقُرْآن", English: "the Qur'an", Transliteration: "Al-qur'aan", Root: "ق ر أ", Type: "noun", Occurrences: 70},
			{Arabic: "آيَة", English: "a sign; a verse", Transliteration: "aayah", Root: "أ ي ي", Type: "noun", Occurrences: 382},
			{Arabic: "أَنْزَلْنَا", English: "We sent down", Transliteration: "anzalnaa", Root: "ن ز ل", Type: "verb", Occurrences: 55},
			{Arabic: "حَدِيث", English: "statement; story", Transliteration: "hadeeth", Root: "ح د ث", Type: "noun", Occurrences: 23},
		},
	},
	{
		Code:        "creation",
		Title:       "Creation",
		Description: "Heaven, earth, day, and the act of creating.",
		Icon:        "🌍",
		Words: []Word{
			{Arabic: "الْأَرْض", English: "the earth", Transliteration: "Al-ard", Root: "أ ر ض", Type: "noun", Occurrences: 461},
			{Arabic: "السَّمَاء", English: "the heaven", Transliteration: "As-samaa'", Root: "س م و", Type: "noun", Occurrences: 310},
			{Arabic: "يَوْم", English: "day", Transliteration: "yawm", Root: "ي و م", Type: "noun", Occurrences: 405},
			{Arabic: "شَيْء", English: "thing", Transliteration: "shay'", Root: "ش ي ء", Type: "noun", Occurrences: 283},
			{Arabic: "خَلَقَ", English: "he created", Transliteration: "khalaqa", Root: "خ ل ق", Type: "verb", Occurrences: 237},
		},
	},
	{
		Code:        "worship-and-remembrance",
		Title:       "Worship & Remembrance",
		Description: "Worshipping, remembering, praying, believing.",
		Icon:        "📿",
		Words: []Word{
			{Arabic: "عَبَدَ", English: "he worshipped", Transliteration: "'abada", Root: "ع ب د", Type: "verb", Occurrences: 142},
			{Arabic: "ذَكَرَ", English: "he remembered", Transliteration: "'akara", Root: "ذ ك ر", Type: "verb", Occurrences: 151},
			{Arabic: "الصَّلَاة", English: "the prayer", Transliteration: "As-salaah", Root: "ص ل و", Type: "noun", Occurrences: 83},
			{Arabic: "اسْم", English: "name", Transliteration: "ism", Root: "س م و", Type: "noun", Occurrences: 39},
			{Arabic: "آمَنُوا", English: "they believed", Transliteration: "aamanoo", Root: "أ م ن", Type: "verb", Occurrences: 258},
		},
	},
	{
		Code:        "everyday-verbs",
		Title:       "Everyday Verbs",
		Description: "Five very common past-tense verbs.",
		Icon:        "🏃",
		Words: []Word{
			{Arabic: "جَاءَ", English: "he came", Transliteration: "jaa'a", Root: "ج ي ء", Type: "verb", Occurrences: 171},
			{Arabic: "فَعَلَ", English: "he did", Transliteration: "fa'ala", Root: "ف ع ل", Type: "verb", Occurrences: 100},
			{Arabic: "جَعَلَ", English: "he made", Transliteration: "ja'ala", Root: "ج ع ل", Type: "verb", Occurrences: 344},
			{Arabic: "فَتَحَ", English: "he opened", Transliteration: "fataha", Root: "ف ت ح", Type: "verb", Occurrences: 25},
			{Arabic: "ضَرَبَ", English: "he struck", Transliteration: "daraba", Root: "ض ر ب", Type: "verb", Occurrences: 50},
		},
	},
	{
		Code:        "prepositions-one",
		Title:       "Prepositions I",
		Description: "The five little words that hold sentences together.",
		Icon:        "🔗",
		Words: []Word{
			{Arabic: "بِ", English: "with; by", Transliteration: "bi", Root: "", Type: "particle", Occurrences: 510},
			{Arabic: "فِي", English: "in; into", Transliteration: "fee", Root: "", Type: "particle", Occurrences: 1684},
			{Arabic: "عَلَى", English: "on; upon", Transliteration: "'alaa", Root: "", Type: "particle", Occurrences: 1423},
			{Arabic: "إِلَى", English: "to; towards", Transliteration: "ilaa", Root: "", Type: "particle", Occurrences: 736},
			{Arabic: "لِ", English: "for", Transliteration: "li", Root: "", Type: "particle", Occurrences: 1361},
		},
	},
	{
		Code:        "prepositions-two",
		Title:       "Prepositions II",
		Description: "From, about, with, near — plus 'or'.",
		Icon:        "🧭",
		Words: []Word{
			{Arabic: "مِنْ", English: "from", Transliteration: "min", Root: "", Type: "particle", Occurrences: 3215},
			{Arabic: "عَنْ", English: "about; away from", Transliteration: "'an", Root: "", Type: "particle", Occurrences: 416},
			{Arabic: "مَعَ", English: "together with", Transliteration: "ma'a", Root: "", Type: "particle", Occurrences: 163},
			{Arabic: "عِنْدَ", English: "in the presence of", Transliteration: "'inda", Root: "", Type: "particle", Occurrences: 197},
			{Arabic: "أَوْ", English: "or", Transliteration: "aw", Root: "", Type: "particle", Occurrences: 280},
		},
	},
	{
		Code:        "negation-and-restriction",
		Title:       "Negation & Restriction",
		Description: "Saying no — and saying 'nothing but'.",
		Icon:        "🚫",
		Words: []Word{
			{Arabic: "لَا", English: "not; no", Transliteration: "laa", Root: "", Type: "particle", Occurrences: 1687},
			{Arabic: "لَمْ", English: "did not", Transliteration: "lam", Root: "", Type: "particle", Occurrences: 384},
			{Arabic: "لَنْ", English: "will never", Transliteration: "lan", Root: "", Type: "particle", Occurrences: 106},
			{Arabic: "إِلَّا", English: "except", Transliteration: "illaa", Root: "", Type: "particle", Occurrences: 664},
			{Arabic: "إِنَّمَا", English: "only", Transliteration: "innamaa", Root: "", Type: "particle", Occurrences: 145},
		},
	},
	{
		Code:        "time-and-condition",
		Title:       "Time & Condition",
		Description: "If, when, and what comes next.",
		Icon:        "⏳",
		Words: []Word{
			{Arabic: "إِنْ", English: "if", Transliteration: "in", Root: "", Type: "particle", Occurrences: 691},
			{Arabic: "لَوْ", English: "had it been", Transliteration: "law", Root: "", Type: "particle", Occurrences: 201},
			{Arabic: "إِذْ", English: "when (in the past)", Transliteration: "i'", Root: "", Type: "particle", Occurrences: 239},
			{Arabic: "إِذَا", English: "when (in the future)", Transliteration: "i'aa", Root: "", Type: "particle", Occurrences: 423},
			{Arabic: "سَوْفَ", English: "soon", Transliteration: "sawfa", Root: "", Type: "particle", Occurrences: 42},
		},
	},
	{
		Code:        "emphasis-and-clauses",
		Title:       "Emphasis & Clauses",
		Description: "Indeed, certainly, that — the connectors of assertion.",
		Icon:        "❗",
		Words: []Word{
			{Arabic: "إِنَّ", English: "indeed", Transliteration: "inna", Root: "", Type: "particle", Occurrences: 1534},
			{Arabic: "أَنْ", English: "that; to", Transliteration: "an", Root: "", Type: "particle", Occurrences: 571},
			{Arabic: "أَنَّ", English: "that (indeed)", Transliteration: "anna", Root: "", Type: "particle", Occurrences: 359},
			{Arabic: "لَقَدْ", English: "certainly", Transliteration: "laqad", Root: "", Type: "particle", Occurrences: 406},
			{Arabic: "شَاءَ", English: "he willed", Transliteration: "shaa'a", Root: "ش ي ء", Type: "verb", Occurrences: 56},
		},
	},
	{
		Code:        "hearing-knowing-doing",
		Title:       "Hearing, Knowing, Doing",
		Description: "Verbs of perception and action, plus what fills the chest.",
		Icon:        "👂",
		Words: []Word{
			{Arabic: "سَمِعَ", English: "he heard", Transliteration: "sami'a", Root: "س م ع", Type: "verb", Occurrences: 98},
			{Arabic: "عَلِمَ", English: "he knew", Transliteration: "'alima", Root: "ع ل م", Type: "verb", Occurrences: 562},
			{Arabic: "عَمِلَ", English: "he acted", Transliteration: "'amila", Root: "ع م ل", Type: "verb", Occurrences: 318},
			{Arabic: "صَدْر", English: "chest", Transliteration: "sadr", Root: "ص د ر", Type: "noun", Occurrences: 44},
			{Arabic: "كَثِير", English: "much; many", Transliteration: "katheer", Root: "ك ث ر", Type: "noun", Occurrences: 63},
		},
	},
	{
		Code:        "mercy-and-forgiveness",
		Title:       "Mercy & Forgiveness",
		Description: "Mercy, peace, sin, and the One who forgives.",
		Icon:        "💚",
		Words: []Word{
			{Arabic: "رَحْمَة", English: "mercy", Transliteration: "rahmah", Root: "ر ح م", Type: "noun", Occurrences: 114},
			{Arabic: "غَفُور", English: "Oft-Forgiving", Transliteration: "ghafoor", Root: "غ ف ر", Type: "noun", Occurrences: 91},
			{Arabic: "الذَّنْب", English: "the sin", Transliteration: "A'-'anb", Root: "ذ ن ب", Type: "noun", Occurrences: 39},
			{Arabic: "سَلَام", English: "peace", Transliteration: "salaam", Root: "س ل م", Type: "noun", Occurrences: 42},
			{Arabic: "طَيِّبَات", English: "good things", Transliteration: "tayyibaat", Root: "ط ي ب", Type: "noun", Occurrences: 46},
		},
	},
	{
		Code:        "this-world-and-the-next",
		Title:       "This World & the Next",
		Description: "The final five: dunya, akhirah, and what awaits.",
		Icon:        "🔥",
		Words: []Word{
			{Arabic: "الدُّنْيَا", English: "the world", Transliteration: "Ad-dunyaa", Root: "د ن و", Type: "noun", Occurrences: 115},
			{Arabic: "الْآخِرَة", English: "the Hereafter", Transliteration: "Al-aakhirah", Root: "أ خ ر", Type: "noun", Occurrences: 115},
			{Arabic: "عَذَاب", English: "punishment", Transliteration: "'a'aab", Root: "ع ذ ب", Type: "noun", Occurrences: 322},
			{Arabic: "النَّار", English: "the Fire", Transliteration: "an-naar", Root: "ن و ر", Type: "noun", Occurrences: 145},
			{Arabic: "عَمَل", English: "deed", Transliteration: "'amal", Root: "ع م ل", Type: "noun", Occurrences: 41},
		},
	},
}
