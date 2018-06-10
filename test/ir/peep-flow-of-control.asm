# Test to demonstrate flow-of-control peephole optimization.

	.data
n:		.word	0

	.text

runtime:
	addi	$sp, $sp, -4
	sw	$ra, 0($sp)

	lw	$ra, 0($sp)
	addi	$sp, $sp, 4
	jr	$ra
	.end runtime


	.globl main
	.ent main
main:
	li	$3, 0		# n -> $3
	# Store dirty variables back into memory
	sw	$3, n

label1:
	# The generated assembly should not contain the label "label2".
	lw	$3, n		# n -> $3
	addi	$3, $3, 1
	# Store dirty variables back into memory
	sw	$3, n
	j	label3

label3:
	li	$2, 10
	syscall
	.end main
