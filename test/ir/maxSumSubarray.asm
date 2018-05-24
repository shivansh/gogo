	.data
a:		.space	40
sum:		.space	40
i:		.word	0
v:		.word	0
v1:		.word	0
v2:		.word	0
maxsum:		.word	0

	.text
	.globl main
	.ent main
main:
	li	$3, 0		# i -> $3
	# Store variables back into memory
	sw	$3, i

loop1:
	lw	$3, i
	bge	$3, 10, exit1	# exit1 -> $0
	li	$v0, 5
	syscall
	move	$7, $v0
	la	$15, a
	sll $s2, $3, 2	# iterator *= 4
	sw	$7, a($s2)	# variable -> array
	addi	$3, $3, 1	# i -> $3
	# Store variables back into memory
	sw	$3, i
	sw	$7, v
	j	loop1

exit1:
	la	$3, a
	lw	$7, 0($3)	# variable <- array
	la	$30, sum
	sw	$7, 0($30)	# variable -> array
	li	$15, 1		# i -> $15
	# Store variables back into memory
	sw	$7, v
	sw	$15, i

loop2:
	lw	$3, i
	bge	$3, 10, exit2	# exit2 -> $0
	sub	$7, $3, 1	# v1 -> $7
	lw	$15, v
	# Store variables back into memory
	sw	$3, i
	sw	$7, v1
	sw	$15, v
	bge	$15, 0, branch1	# branch1 -> $0

	la	$3, a
	lw	$7, i
	sll	$s2, $7, 2	# iterator *= 4
	lw	$30, a($s2)	# variable <- array
	la	$15, sum
	sll $s2, $7, 2	# iterator *= 4
	sw	$30, sum($s2)	# variable -> array
	addi	$7, $7, 1	# i -> $7
	# Store variables back into memory
	sw	$7, i
	sw	$30, v
	j	loop2

branch1:
	la	$3, a
	lw	$7, i
	sll	$s2, $7, 2	# iterator *= 4
	lw	$30, a($s2)	# variable <- array
	sub	$15, $7, 1	# v1 -> $15
	la	$16, sum
	sll	$s2, $15, 2	# iterator *= 4
	lw	$8, sum($s2)	# variable <- array
	add	$30, $30, $8	# v -> $30
	sll $s2, $7, 2	# iterator *= 4
	sw	$30, sum($s2)	# variable -> array
	addi	$7, $7, 1	# i -> $7
	# Store variables back into memory
	sw	$7, i
	sw	$8, v2
	sw	$15, v1
	sw	$30, v
	j	loop2

exit2:
	la	$3, sum
	lw	$7, 0($3)	# variable <- array
	li	$30, 1		# i -> $30
	# Store variables back into memory
	sw	$7, maxsum
	sw	$30, i

loop3:
	lw	$3, i
	bge	$3, 10, exit3	# exit3 -> $0
	la	$7, sum
	sll	$s2, $3, 2	# iterator *= 4
	lw	$15, sum($s2)	# variable <- array
	lw	$30, maxsum
	# Store variables back into memory
	sw	$3, i
	sw	$15, v
	sw	$30, maxsum
	bge	$30, $15, branch2	# branch2 -> $0

	lw	$3, v
	move	$7, $3		# maxsum -> $7
	lw	$30, i
	addi	$30, $30, 1	# i -> $30
	# Store variables back into memory
	sw	$3, v
	sw	$7, maxsum
	sw	$30, i
	j	loop3

branch2:
	lw	$3, i
	addi	$3, $3, 1	# i -> $3
	# Store variables back into memory
	sw	$3, i
	j	loop3

exit3:
	li	$v0, 1
	lw	$3, maxsum
	move	$a0, $3
	syscall
	# Store variables back into memory
	sw	$3, maxsum
	li	$v0, 10
	syscall
	.end main
