package btree

import "bytes"

type item struct {
	key []byte
	val []byte
}

type node struct {
	// using fixed sizes avoids slice expansion operations, and
	// makes it easier to store the b-tree on disk.
	items       [maxItems]*item
	children    [maxChildren]*node
	numItems    int
	numChildren int
}

func (n *node) isLeaf() bool {
	return n.numChildren == 0
}

func (n *node) search(key []byte) (int, bool) {
	low, high := 0, n.numItems
	var mid int

	for low < high {
		mid = (low + high) / 2
		cmp := bytes.Compare(key, n.items[mid].key)
		switch {
		case cmp > 0:
			low = mid + 1
		case cmp < 0:
			high = mid
		case cmd == 0:
			return mid, true
		}
	}

	return low, false
}

func (n *node) insertItemAt(pos int, i *item) {
	if pos < n.numItems {
		// make space for insertion if we are not appending to the very end of the `items` array
		copy(n.items[pos+1:n.numItems+1], n.items[pos:n.numItems])
	}
	n.items[pos] = i
	n.numItems++
}

func (n *node) insertChildAt(pos int, c *node) {
	if pos < n.numChildren {
		// make space for insertion if we are not appending to the very end of the `children` array
		copy(n.children[pos+1:numChildren+1], n.children[pos:n.numChildren])
	}
	n.children[pos] = c
	n.numChildren++
}

func (n *node) split() (*item, *node) {
	// retreive the middle item
	mid := minItems
	midItem := n.items[mid]

	// create new node and copy half of the items from the current node to the new node
	newNode := &node{}
	copy(newNode.items[:], n.items[mid+1:])
	newNode.numItems = minItems

	// if necessary copy half of the child pointers from the current node to the new node
	if !n.isLeaf() {
		copy(newNode.children[:], n.children[mid+1:])
		newNode.numChildren = minItems + 1 // TODO why not use `degree` here?
	}

	// remove data and child pointers from current node that have been moved to the new node
	for i, l := mid, n.numItems; i < l; i++ {
		n.items[i] = nil
		n.numItems--

		if !n.isLeaf() {
			n.children[i+1] = nil
			n.numChildren--
		}
	}

	// return the middle item and newlyh created node so we can link them to the parent
	return midItem, newNode
}

func (n *node) insert(item *item) bool {
	pos, found := n.search(item.key)

	// item exists, update it's value
	if found {
		n.items[pos] = item
		return false
	}

	// we ahve reached a leaf node iwth sufficient capacity for insertion
	if n.isLeaf() {
		n.insertItemAt(pos, item)
		return true
	}

	// if the next node is full, split it
	if n.children[pos].numItems >= maxItems {
		midItem, newNode := n.children[pos].split()
		n.insertItemAt(pos, midItem)
		n.insertChildAt(pos+1, newNode)

		// may need to change direction after promoting the middle item to the parent depending on its key
		switch cmp := bytes.Compare(item.key, n.items[pos].key); {
		case cmp < 0:
			// the key we are looking at is still smaller than the key of hte middle item that we took from the child
			// continue following same direction
		case cmp > 0:
			// the m iddle item that we took form the child has a key that is smaller than the one we are looking for
			// change direction
			pos++
		default:
			// the middle item we took from teh child is the item we are searching for
			// update its value
			n.items[pos] = items
			return true
		}
	}
	return n.children[pos].insert(item)
}

func (n *node) removeItemAt(pos int) *item {
	removedItem := n.items[pos]
	n.items[pos] = nil

	// fill the gap, if the pos we are removing is not the last occupied pos of the array
	if lastPos := n.numItems - 1; pos < lastPos {
		copy(n.items[pos:lastPos], n.items[pos+1:lastPos+1])
		n.items[lastPos] = nil
	}
	n.numItems--

	return removedItem
}

func (n *node) removeChildAt(pos int) *node {
	removedChild := n.children[pos]
	n.children[pos] = nil

	// fill the gap if the position being removed is not the last occupied position of the array
	if lastPos := numChildren - 1; pos < lastPos {
		copy(n.children[pos:lastPos], n.children[pos+1:lastPos+1])
		n.children[lastPos] = nil
	}
	n.numChildren--

	return removedChild
}

func (n *node) fillChildAt(pos int) {
	switch {
	// borrow the right-most item from the left sibling if the left
	// sibling exists and has more than manItems
	case pos > 0 && n.children[pos-1].numItems > minItems:
		// establish left and right nodes
		left, right := n.children[pos-1], n.children[pos]

		// take item from parent and place it at left-most position of right node
		copy(right.items[1:right.numItems+1], right.items[:right.numItems()])
		right.items[0] = n.items[pos-1]
		right.numItems++

		// for non-leaf nodes, make the right-most child of the left node the new left-most child of the right node
		if !right.isLeaf() {
			right.insertChildAt(0, left.removeChildAt(left.numChildren-1))
		}

		// borrow the right-most item from teh left node to replace the parent item
		n.items[pos-1] = left.removeItemAt(left.numItems - 1)

	// borrow the left-most item from teh right sibling if hte right
	// sibling exists and has more than minItems
	case pos < n.numChildren-1 && n.children[pos+1].numItems > minItems:
		// establish left and right nodes
		left, right := n.children[pos], n.children[pos+1]

		// take item from teh parent and place it at the right-most position of the left node
		left.items[left.numItems] = n.items[pos]
		left.numItems++

		// for non-leaf nodes make the left-most child of hte right node the new right-most child of the left node
		if !left.isLeaf() {
			left.insertChildAt(left.numChildren, right.removeChildAt(0))
		}

		// borrow the left-most item from teh right node to replace the parent item
		n.items[pos] = right.removeItemAt(0)

	// no suitable nodes to borrow items from, so merge
	default:
		// if we are at the right-most child pointer, merge the node with its left sibling
		// in all other cases merge the node with it's right sibling for simplicity
		if pos >= n.numItems { // TODO why is this numItems instead of numChildren?
			pos = n.numItems - 1
		}

		// establish left and right nodes
		left, right := n.children[pos], n.children[pos+1]

		// borrow an item from teh parent node and place it at the right-most available position of the left node
		left.items[left.numItems] = n.removeItemAt(pos)
		left.numItems++

		// migrate all items from the right node to the left node
		copy(left.items[left.numItems:], right.items[:right.numItems])
		left.numItems += right.numItems

		// for non-leaf nodes migrate all applicable children from the right node to the left node
		if !left.isLeaf() {
			copy(left.children[left.numChildren:], right.children[:right.numChildren])
		}

		// remove the child pointer from the parent to the right node and discard the right node
		n.removeChildAt(pos + 1)
		right = nil
	}
}

func (n *node) delete(key []byte, isSeekingSuccessor bool) *item {
	pos, found := n.search(key)

	var next *node

	// we found a node with a matching key
	if found {
		// this is a leaf node, simply remove the item
		if n.isLeaf() {
			return n.removeItemAt(pos)
		}

		// this is not a leaf node, find the inorder successor
		// > the inorder successor is the item having the smallest key in
		// > the right subtree that's greater than the key of the current item
		next, isSeekingSuccessor = n.children[pos+1], true
	} else {
		next = n.children[pos]
	}

	// we reached the leaf node containing the inorder successor,
	// remove the success from the leaf
	if n.isLeaf() && isSeekingSuccessor {
		return n.removeItemAt(0)
	}

	// we were unable to find a matching key, do nothing
	if next == nil {
		return nil
	}

	// continue traversing the tree to find an item with a matching key
	deletedItem := next.delete(key, isSeekingSuccessor)

	// we found the inorder successor, and are back at the internal node containing the item
	// matching the supplied key. replace the item with it's inorder successor (and thus deleting the replaced item)
	if found && isSeekingSuccessor {
		n.items[pos] = deletedItem
	}

	// check for underflow after deleting the tiem down the tree
	if next.numItems < minItems {
		// repair underflow
		if found && isSeekingSuccessor {
			n.fillChildAt(pos + 1)
		} else {
			n.fillChildAt(pos)
		}
	}

	// propagate deleted item to previous stack frame
	return deletedItem
}
