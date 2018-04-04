	.data
a:	.word	0
t0:	.word	0
t1:	.word	0

	.text


	.globl main
	.ent main
main:
	li $t1, 1		# a -> $t1
	# Store variables back into memory
	sw $t1, a

	lw $t1, a
	beq $t1, 1, l1		# l1 -> $t0
	# Store variables back into memory
	sw $t1, a

	j l2

l1:
	lw $t1, a
	addi $t4, $t1, 1		# t0 -> $t4
	move $t1, $t4		# a -> $t1
	# Store variables back into memory
	sw $t1, a
	sw $t4, t0

	j l0

l2:
	lw $t1, a
	addi $t4, $t1, 4		# t1 -> $t4
	move $t1, $t4		# a -> $t1
	# Store variables back into memory
	sw $t4, t1
	sw $t1, a

	j l0

l0:
	li $v0, 10
	syscall
	.end main
