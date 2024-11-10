package main

import (
	"fmt"
	"sort"
)

// graph - DP - nPath路径

// NSP 一个简单的分词器
type NSP struct {
	// nPath 最优路径
	nPath int
	// dict 词典
	dict map[string]struct{}
	// maxWordLen 最大word长度，句子切分后的最大长度
	maxWordLen int
}

// VNode 图起点
type VNode struct {
	// vNodeNum 在list中的编号
	vNodeNum int
	// eNode vNode关联的ENode
	eNode *ENode
}

// ENode 边
type ENode struct {
	// endPoint word末尾在句子中的结束位置，从1开始
	endPoint int
	// weight 边的权重
	weight int
	// word 边上的词
	word string
	// next 边的下一条边
	next *ENode
}

// Token 句子切词的结果
type Token struct {
	// token在句子中的起始位置
	start int
	// token长度
	len   int
	token string
}

type PathCollection struct {
	// weightToPaths
	weightToPaths map[int][]*Path
}

type Path struct {
	// pathWeight 路径上的权重总和
	pathWeight int
	// wordPath word的path路径，以“\t”分割
	wordPath string
}

func NewNSP(nPath int) *NSP {
	dict := map[string]struct{}{}

	return &NSP{
		nPath:      nPath,
		dict:       dict,
		maxWordLen: 6,
	}
}

func (nsp *NSP) SetDict(dict map[string]struct{}) {
	if len(dict) > 0 {
		nsp.dict = dict
	}
}

func (nsp *NSP) createGraph(sentence string) []*VNode {
	sentenceRuneSlice := []rune(sentence)
	length := len(sentenceRuneSlice)
	// 1. 构建图
	graph := make([]*VNode, 0, length)
	for i := 0; i < length; i++ {
		vNode := &VNode{
			vNodeNum: i,
			eNode:    nil,
		}
		eNode := &ENode{
			endPoint: i + 1,
			weight:   1,
			word:     string(sentenceRuneSlice[i : i+1]),
			next:     nil,
		}

		vNode.eNode = eNode
		graph = append(graph, vNode)
	}
	// 添加一个尾节点
	//（paths: List<PathCollection>含义：paths[0]表示到sentenceRuneSlice[0]的路径集合）
	//                                 paths[1]表示到sentenceRuneSlice[0]的路径集合
	//                                 这里的0，1..表示eNode的endPoint位置，0是需要特殊初始化的
	//                                 1..这些是递推来的，求的是paths[len(表示到sentenceRuneSlice)],所以要有一个尾节点
	tailVNode := &VNode{
		vNodeNum: length,
		eNode:    nil,
	}
	graph = append(graph, tailVNode)

	//
	tokens := nsp.splitSentence(sentenceRuneSlice)
	for _, token := range tokens {
		start := token.start
		vNode := graph[start]

		// 新建一条边并添加到VNode的边的末尾
		eNode := &ENode{
			endPoint: start + token.len, // endPoint从1开始，所以这里是start+token.len
			weight:   1,
			word:     token.token,
			next:     nil,
		}

		eNodeTemp := vNode.eNode
		for eNodeTemp.next != nil {
			eNodeTemp = eNodeTemp.next
		}
		eNodeTemp.next = eNode
	}

	return graph
}

func (nsp *NSP) GetNPath(sentence string) []string {
	// 创建图
	graph := nsp.createGraph(sentence)
	// DP - 计算
	pathCollection := nsp.computePath(graph)
	// getNPath
	paths := nsp.getNPath(pathCollection)

	return paths
}

// splitSentence 切分句子为词，最长为nsp.MaxWordLen，最短为2
func (nsp *NSP) splitSentence(sentence []rune) []Token {
	tokens := make([]Token, 0, 10)
	for i := 0; i < len(sentence); i++ {
		wordLen := len(sentence) - i
		if wordLen > nsp.maxWordLen {
			wordLen = nsp.maxWordLen
		}

		// word的长度最小为2
		for ; wordLen >= 2; wordLen-- {
			subWord := string(sentence[i : i+wordLen])
			_, ok := nsp.dict[subWord]
			if ok {
				token := Token{
					start: i,
					len:   wordLen,
					token: subWord,
				}
				tokens = append(tokens, token)
			}
		}

	}

	return tokens
}

// computePath 计算路径 - DP
func (nsp *NSP) computePath(vNodeList []*VNode) *PathCollection {
	pathCollections := make([]*PathCollection, 0, len(vNodeList))
	for i := 0; i < len(vNodeList); i++ {
		pathCollection := &PathCollection{
			weightToPaths: map[int][]*Path{},
		}
		pathCollections = append(pathCollections, pathCollection)
	}

	// 求pathCollections[len(pathCollections)-1]

	// 初始化pathCollections[1]
	eNodeTemp := vNodeList[0].eNode
	for eNodeTemp != nil {
		// 构架了一个到达eNodeTemp.endPoint的新路径
		path := &Path{
			pathWeight: eNodeTemp.weight,
			wordPath:   eNodeTemp.word,
		}

		pathCollectionTemp := pathCollections[eNodeTemp.endPoint] // eNodeTemp.endPoint的路径集合
		// 收集路径，相同weight的放在同一组（方便最终求nPath）
		weightPaths, ok := pathCollectionTemp.weightToPaths[path.pathWeight]
		if ok {
			pathCollectionTemp.weightToPaths[path.pathWeight] = append(weightPaths, path)
		} else {
			pathCollectionTemp.weightToPaths[path.pathWeight] = []*Path{path}
		}

		eNodeTemp = eNodeTemp.next
	}

	// 开始递推
	for i := 1; i < len(vNodeList); i++ {
		curNode := vNodeList[i]
		curNodePathCollection := pathCollections[curNode.vNodeNum]
		// 对于vNodeList[i]的所有边，均以curNodePathCollection开始递推
		// curNodePathCollection表示到句子i位置的路径，vNodeList[i]的所有边到句子的位置均大于i

		// 从 i 到 eNode.endPoint 的路径递推
		// i的位置加上eNode.word就到了eNode.endPoint的位置
		// 比如vNode=2->中(边endpoint=3)->中国特色(边endpoint=6)->中国(边endpoint=4)
		// 比如eNode=中国特色,vNode=2(“设”的位置),设+中国特色=色的endPoint)

		eNode := curNode.eNode
		for eNode != nil {
			curENodePathCollection := pathCollections[eNode.endPoint]
			for _, curNodePaths := range curNodePathCollection.weightToPaths {
				for _, curNodePath := range curNodePaths {
					path := &Path{
						pathWeight: curNodePath.pathWeight + eNode.weight,
						wordPath:   curNodePath.wordPath + "-" + eNode.word,
					}

					_, ok := curENodePathCollection.weightToPaths[path.pathWeight]
					if ok {
						curENodePathCollection.weightToPaths[path.pathWeight] = append(curENodePathCollection.
							weightToPaths[path.pathWeight], path)
					} else {
						curENodePathCollection.weightToPaths[path.pathWeight] = []*Path{path}
					}
				}
			}
			eNode = eNode.next
		}
	}

	return pathCollections[len(pathCollections)-1]
}

func (nsp *NSP) getNPath(pathCollection *PathCollection) []string {
	weightToPaths := pathCollection.weightToPaths

	// weight从小到大排序
	temp := make([]int, 0, len(weightToPaths))
	for k := range weightToPaths {
		temp = append(temp, k)
	}
	sort.Ints(temp)

	weightToPathsTemp := map[int][]*Path{}

	minLength := len(temp)
	if minLength > nsp.nPath {
		minLength = nsp.nPath
	}
	for i := 0; i < minLength; i++ {
		weightToPathsTemp[temp[i]] = weightToPaths[temp[i]]
	}

	result := make([]string, 0, 10)
	for _, paths := range weightToPathsTemp {
		for _, path := range paths {
			result = append(result, path.wordPath)
		}
	}

	return result
}

func main() {
	// 创建nsp
	nsp := NewNSP(10)

	// 设置词典
	nspDict := map[string]struct{}{}
	nspDict["围城"] = struct{}{}
	nspDict["故事"] = struct{}{}
	nspDict["1920"] = struct{}{}
	nspDict["1940"] = struct{}{}
	nspDict["年代"] = struct{}{}
	nspDict["主角"] = struct{}{}
	nspDict["方泓渐"] = struct{}{}
	nspDict["中国"] = struct{}{}
	nspDict["男方"] = struct{}{}
	nspDict["乡绅"] = struct{}{}
	nspDict["中国男方"] = struct{}{}
	nspDict["家庭"] = struct{}{}
	nspDict["乡绅家庭"] = struct{}{}
	nspDict["迫于"] = struct{}{}
	nspDict["青年"] = struct{}{}
	nspDict["青年人"] = struct{}{}
	nspDict["压力"] = struct{}{}
	nspDict["家庭压力"] = struct{}{}
	nspDict["同乡"] = struct{}{}
	nspDict["周家"] = struct{}{}
	nspDict["女子"] = struct{}{}
	nspDict["女子定亲"] = struct{}{}
	nspDict["定亲"] = struct{}{}

	nsp.SetDict(nspDict)

	// 构建图并获取nPath路径
	sentence := "围城故事发生于1920到1940年代。主角方鸿渐是个从中国南方乡绅家庭走出的青年人，迫于家庭压力与同乡周家女子订亲"
	paths := nsp.GetNPath(sentence)

	for _, path := range paths {
		fmt.Println(path)
	}

	fmt.Println("len(paths): ", len(paths)) // OUTPUT 37346
}
