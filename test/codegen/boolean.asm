	.data
a.0:		.word	0
b.1:		.word	0
t0:		.word	0
t1:		.word	0
t2:		.word	0

	.text

	.globl main
	.ent main
main:
	li	$3, 1		# a.0 -> $3
	li	$5, 2		# b.1 -> $5
	# Store dirty variables back into memory
	sw	$3, a.0
	sw	$5, b.1
	bge	$3, $5, l0

	li	$3, 1		# t0 -> $3
	# Store dirty variables back into memory
	sw	$3, t0
	j	l1

l0:
	li	$3, 0		# t0 -> $3
	# Store dirty variables back into memory
	sw	$3, t0

l1:
	lw	$3, a.0		# a.0 -> $3
	lw	$5, b.1		# b.1 -> $5
	bge	$3, $5, l2

	li	$3, 1		# t1 -> $3
	# Store dirty variables back into memory
	sw	$3, t1
	j	l3

l2:
	li	$3, 0		# t1 -> $3
	# Store dirty variables back into memory
	sw	$3, t1

l3:
	lw	$3, t0		# t0 -> $3
	beq	$3, 0, l5

	lw	$3, t1		# t1 -> $3
	beq	$3, 0, l5

	li	$3, 1		# t2 -> $3
	# Store dirty variables back into memory
	sw	$3, t2
	j	l4

l5:
	li	$3, 0		# t2 -> $3
	# Store dirty variables back into memory
	sw	$3, t2

l4:
	lw	$3, t2		# t2 -> $3
	blt	$3, 1, l6

	li	$2, 1
	lw	$3, a.0		# a.0 -> $3
	move	$4, $3
	syscall

l6:
	li	$2, 10
	syscall
	.end main
