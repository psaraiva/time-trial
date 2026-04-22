package generator

// adjectives and objects are used by generateStringFunny to produce absurd,
// human-friendly names by combining unrelated words.
// Format: <adjective>_<object>  e.g. "angry_spoon", "sleepy_bucket"
var adjectives = []string{
	"admiring", "adoring", "affectionate", "agitated", "amazing",
	"angry", "awesome", "blissful", "bold", "brave",
	"busy", "charming", "clever", "compassionate", "competent",
	"condescending", "confident", "crazy", "creative", "curious",
	"dazzling", "determined", "distracted", "eager", "ecstatic",
	"elastic", "elated", "elegant", "eloquent", "epic",
	"exciting", "fervent", "festive", "flamboyant", "focused",
	"friendly", "frosty", "funny", "gallant", "gifted",
	"goofy", "gracious", "great", "happy", "hardcore",
	"heuristic", "hopeful", "hungry", "infallible", "inspiring",
	"intelligent", "jolly", "jovial", "keen", "kind",
	"laughing", "lucid", "magical", "modest", "mystifying",
	"naughty", "nervous", "nice", "nifty", "nostalgic",
	"objective", "optimistic", "peaceful", "pedantic", "pensive",
	"practical", "priceless", "quirky", "recursing", "relaxed",
	"reverent", "romantic", "sad", "serene", "sharp",
	"silly", "sleepy", "stoic", "strange", "stupefied",
	"suspicious", "sweet", "tender", "thirsty", "trusting",
	"unruffled", "upbeat", "vibrant", "vigilant", "vigorous",
	"wizardly", "wonderful", "xenodochial", "youthful", "zealous",
}

// objects are everyday items whose combination with adjectives produces
// absurd and memorable names.
var objects = []string{
	"anchor", "anvil", "apron", "axe", "bag",
	"ball", "barrel", "basket", "bathtub", "bell",
	"belt", "blanket", "bolt", "boot", "bottle",
	"bowl", "box", "brick", "broom", "brush",
	"bucket", "button", "cabinet", "candle", "chair",
	"clock", "comb", "cord", "cup", "curtain",
	"desk", "door", "drawer", "drum", "envelope",
	"eraser", "fan", "fence", "fork", "frame",
	"fridge", "funnel", "gear", "glove", "hammer",
	"handle", "hat", "hook", "hose", "jar",
	"jug", "kettle", "key", "knob", "ladder",
	"lamp", "latch", "ladle", "lever", "lid",
	"lock", "map", "mirror", "mop", "mug",
	"nail", "needle", "net", "notebook", "pan",
	"pencil", "pillow", "pin", "pipe", "pitcher",
	"plate", "pliers", "plug", "pocket", "pot",
	"pump", "rack", "rope", "ruler", "saw",
	"screw", "shelf", "shovel", "sink", "soap",
	"sock", "spatula", "spoon", "stamp", "stool",
	"string", "switch", "table", "tape", "towel",
	"tray", "tube", "valve", "vase", "wallet",
	"wheel", "whisk", "window", "wire", "wrench",
}
