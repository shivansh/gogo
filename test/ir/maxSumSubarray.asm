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
	bge	$3, 10, exit1
	li	$2, 5
	syscall
	move	$5, $2
	la	$6, a
	sll $s2, $3, 2	# iterator *= 4
	sw	$5, a($s2)	# variable -> array
	addi	$3, $3, 1	# i -> $3
	# Store variables back into memory
	sw	$3, i
	sw	$5, v
	j	loop1

exit1:
	la	$3, a
	lw	$5, 0($3)	# variable <- array
	la	$3, sum
	sw	$5, 0($3)	# variable -> array
	li	$3, 1		# i -> $3
	# Store variables back into memory
	sw	$3, i
	sw	$5, v

loop2:
	lw	$3, i
	bge	$3, 10, exit2
	sub	$5, $3, 1	# v1 -> $5
	sw	$5, v1		# spilled v1, freed $5
	lw	$5, v
	# Store variables back into memory
	sw	$3, i
	sw	$5, v
	bge	$5, 0, branch1

	la	$3, a
	lw	$5, i
	sll	$s2, $5, 2	# iterator *= 4
	lw	$6, a($s2)	# variable <- array
	la	$3, sum
	sll $s2, $5, 2	# iterator *= 4
	sw	$6, sum($s2)	# variable -> array
	addi	$5, $5, 1	# i -> $5
	# Store variables back into memory
	sw	$5, i
	sw	$6, v
	j	loop2

branch1:
	la	$3, a
	lw	$5, i
	sll	$s2, $5, 2	# iterator *= 4
	lw	$6, a($s2)	# variable <- array
	sub	$3, $5, 1	# v1 -> $3
	la	$7, sum
	sll	$s2, $3, 2	# iterator *= 4
	lw	$8, sum($s2)	# variable <- array
	add	$6, $6, $8	# v -> $6
	sll $s2, $5, 2	# iterator *= 4
	sw	$6, sum($s2)	# variable -> array
	addi	$5, $5, 1	# i -> $5
	# Store variables back into memory
	sw	$3, v1
	sw	$5, i
	sw	$6, v
	sw	$8, v2
	j	loop2

exit2:
	la	$3, sum
	lw	$5, 0($3)	# variable <- array
	li	$3, 1		# i -> $3
	# Store variables back into memory
	sw	$3, i
	sw	$5, maxsum

loop3:
	lw	$3, i
	bge	$3, 10, exit3
	la	$5, sum
	sll	$s2, $3, 2	# iterator *= 4
	lw	$6, sum($s2)	# variable <- array
	lw	$5, maxsum
	# Store variables back into memory
	sw	$3, i
	sw	$5, maxsum
	sw	$6, v
	bge	$5, $6, branch2

	lw	$3, v
	move	$5, $3		# maxsum -> $5
	sw	$3, v		# spilled v, freed $3
	lw	$3, i
	addi	$3, $3, 1	# i -> $3
	# Store variables back into memory
	sw	$3, i
	sw	$5, maxsum
	j	loop3

branch2:
	lw	$3, i
	addi	$3, $3, 1	# i -> $3
	# Store variables back into memory
	sw	$3, i
	j	loop3

exit3:
	li	$2, 1
	lw	$3, maxsum
	move	$4, $3
	syscall
	# Store variables back into memory
	sw	$3, maxsum
	li	$2, 10
	syscall
	.end main
