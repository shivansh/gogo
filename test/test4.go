package main

import "fmt"

type Tree struct {
	data   float64
	filled bool
	left   *Tree
	right  *Tree
	parent *Tree
}

func (tree *Tree) Insert(key float64) {
	if !tree.filled {
		tree.data = key
		tree.filled = true
	} else {
		var temp *Tree = tree
		for {
			if temp.data > key {
				if temp.left == nil {
					temp.left = new(Tree)
					temp.left.data = key
					temp.left.parent = temp
					temp.left.filled = true
					break
				} else {
					temp = temp.left
				}
			} else if temp.data < key {
				if temp.right == nil {
					temp.right = new(Tree)
					temp.right.data = key
					temp.right.parent = temp
					temp.right.filled = true
					break
				} else {
					temp = temp.right
				}
			} else {
				fmt.Println("This value is already present in the set")
				break
			}
		}
	}
	return
}

func (tree *Tree) Search(key float64) (location *Tree, present bool) {
	if tree.filled == false {
		present = false
		location = nil
	} else {
		var temp *Tree = tree
		for {
			if temp.data == key {
				location = temp
				present = true
				break
			} else if key > temp.data {
				temp = temp.right
				if temp == nil {
					location = nil
					present = false
					break
				}
			} else {
				temp = temp.left
				if temp == nil {
					location = nil
					present = false
					break
				}
			}
		}
	}
	return
}

func (tree *Tree) Delete(key float64) {
	location, present := tree.Search(key)
	if !present {
		fmt.Println("This value is not present in the set")
	} else {
		if location.left == nil && location.right == nil {
			if location.parent.left == location {
				location.parent.left = nil
			} else {
				location.parent.right = nil
			}
		} else if location.right == nil {
			if location.parent.right == location {
				location.parent.right = location.left
				location.left.parent = location.parent
			} else {
				location.parent.left = location.left
				location.left.parent = location.parent
			}
			location = nil
		} else if location.left == nil {
			if location.parent.right == location {
				location.parent.right = location.right
				location.right.parent = location.parent
			} else {
				location.parent.left = location.right
				location.right.parent = location.parent
			}
			location = nil
		} else {
			new_location := location.right
			for new_location.left != nil {
				new_location = new_location.left
			}
			location.data = new_location.data
			new_location = nil
		}
	}
	return
}

func main() {
	var tree Tree
	tree.Insert(50)
	tree.Insert(20)
	tree.Insert(70)
	tree.Insert(10)
	tree.Insert(30)
	tree.Insert(60)
	tree.Insert(80)
	tree.Delete(50)
	_, present := tree.Search(50)
	fmt.Println(present)
}
