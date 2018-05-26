# Test to demonstrate flow-of-control peephole optimization.

	.data
n:		.word	0

	.text


	.globl main
	.ent main
main:
	li	$3, 0		# n -> $3
	# Store variables back into memory
	sw	$3, n

label1:
	# The generated assembly should not contain the label "label2".
	lw	$3, n
	addi	$3, $3, 1	# n -> $3
	# Store variables back into memory
	sw	$3, n
	j	label3

label3:
	li	$2, 10
	syscall
	.end main
